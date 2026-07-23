package http

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
	"time"

	redispkg "github.com/niflaot/pixels/pkg/redis"
)

const (
	// ssoIdempotencyHeader carries the caller-controlled retry key.
	ssoIdempotencyHeader = "Idempotency-Key"
	// ssoReplayHeader identifies a replayed ticket response.
	ssoReplayHeader = "X-Idempotent-Replay"
	// ssoIdempotencyPrefix isolates SSO request records from ticket records.
	ssoIdempotencyPrefix = "pixels:sso:create:"
	// ssoPendingTTL bounds abandoned ticket claims.
	ssoPendingTTL = 90 * time.Second
)

var (
	// errSSOIdempotencyConflict reports a key reused for different input.
	errSSOIdempotencyConflict = errors.New("sso idempotency key conflict")
	// errSSOIdempotencyPending reports concurrent work for one key.
	errSSOIdempotencyPending = errors.New("sso idempotent request is pending")
)

// ssoIdempotencyRecord is the replayable Redis representation.
type ssoIdempotencyRecord struct {
	// State is pending or complete.
	State string `json:"state"`
	// RequestHash binds the key to one request body.
	RequestHash string `json:"requestHash"`
	// Response stores the completed ticket response.
	Response *CreateSSOTicketResponse `json:"response,omitempty"`
}

// ssoIdempotencyStore coordinates retry-safe ticket creation.
type ssoIdempotencyStore struct {
	// client stores claims beside the one-time tickets.
	client *redispkg.Client
}

// claim creates a pending claim or returns an existing record.
func (store ssoIdempotencyStore) claim(ctx context.Context, key string, request CreateSSOTicketRequest) (ssoIdempotencyRecord, bool, error) {
	key = strings.TrimSpace(key)
	if key == "" || len(key) > 128 {
		return ssoIdempotencyRecord{}, false, errSSOIdempotencyConflict
	}
	payload, err := json.Marshal(request)
	if err != nil {
		return ssoIdempotencyRecord{}, false, err
	}
	digest := sha256.Sum256(payload)
	hash := hex.EncodeToString(digest[:])
	record := ssoIdempotencyRecord{State: "pending", RequestHash: hash}
	recordPayload, err := json.Marshal(record)
	if err != nil {
		return ssoIdempotencyRecord{}, false, err
	}
	created, err := store.client.SetIfAbsent(ctx, ssoIdempotencyPrefix+key, recordPayload, ssoPendingTTL)
	if err != nil || created {
		return record, created, err
	}
	existingPayload, found, err := store.client.Find(ctx, ssoIdempotencyPrefix+key)
	if err != nil {
		return ssoIdempotencyRecord{}, false, err
	}
	if !found {
		return ssoIdempotencyRecord{}, false, errSSOIdempotencyPending
	}
	var existing ssoIdempotencyRecord
	if err := json.Unmarshal(existingPayload, &existing); err != nil {
		return ssoIdempotencyRecord{}, false, err
	}
	if existing.RequestHash != hash {
		return ssoIdempotencyRecord{}, false, errSSOIdempotencyConflict
	}
	return existing, false, nil
}

// complete stores a ticket response only for the ticket's remaining lifetime.
func (store ssoIdempotencyStore) complete(ctx context.Context, key string, response CreateSSOTicketResponse, ttl time.Duration) error {
	record := ssoIdempotencyRecord{State: "complete", Response: &response}
	payload, err := json.Marshal(record)
	if err != nil {
		return err
	}
	// Preserve the request hash written by claim before replacing the record.
	existingPayload, found, err := store.client.Find(ctx, ssoIdempotencyPrefix+strings.TrimSpace(key))
	if err != nil {
		return err
	}
	if found {
		var existing ssoIdempotencyRecord
		if err := json.Unmarshal(existingPayload, &existing); err != nil {
			return err
		}
		record.RequestHash = existing.RequestHash
		payload, err = json.Marshal(record)
		if err != nil {
			return err
		}
	}
	return store.client.Set(ctx, ssoIdempotencyPrefix+strings.TrimSpace(key), payload, ttl)
}

// release removes a failed ticket claim so a retry may proceed.
func (store ssoIdempotencyStore) release(ctx context.Context, key string) error {
	return store.client.Delete(ctx, ssoIdempotencyPrefix+strings.TrimSpace(key))
}
