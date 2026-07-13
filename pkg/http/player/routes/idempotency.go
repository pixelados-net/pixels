package routes

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/niflaot/pixels/pkg/redis"
)

const (
	// idempotencyPrefix namespaces administrative player creation records.
	idempotencyPrefix = "pixels:admin:players:create:"
	// idempotencyTTL keeps completed registration results replayable.
	idempotencyTTL = 24 * time.Hour
)

var (
	// errIdempotencyConflict reports one key reused with different input.
	errIdempotencyConflict = errors.New("idempotency key reused with different input")
	// errIdempotencyPending reports concurrent work for the same request.
	errIdempotencyPending = errors.New("idempotent request is still in progress")
)

// idempotencyStore coordinates replayable player creation requests.
type idempotencyStore struct {
	// client persists request state in Redis.
	client *redis.Client
}

// newIdempotencyStore creates an administrative player idempotency store.
func newIdempotencyStore(client *redis.Client) idempotencyStore {
	return idempotencyStore{client: client}
}

// claim creates a pending request or returns its existing record.
func (store idempotencyStore) claim(ctx context.Context, key string, request CreateRequest) (idempotencyRecord, bool, error) {
	normalizedKey := strings.TrimSpace(key)
	if normalizedKey == "" || len(normalizedKey) > 128 {
		return idempotencyRecord{}, false, errIdempotencyConflict
	}

	hash, err := requestHash(request)
	if err != nil {
		return idempotencyRecord{}, false, err
	}
	record := idempotencyRecord{State: "pending", RequestHash: hash, Username: strings.TrimSpace(request.Username)}
	payload, err := json.Marshal(record)
	if err != nil {
		return idempotencyRecord{}, false, err
	}

	created, err := store.client.SetIfAbsent(ctx, store.key(normalizedKey), payload, idempotencyTTL)
	if err != nil {
		return idempotencyRecord{}, false, err
	}
	if created {
		return record, true, nil
	}

	existing, found, err := store.find(ctx, normalizedKey)
	if err != nil {
		return idempotencyRecord{}, false, err
	}
	if !found || existing.RequestHash != hash {
		return idempotencyRecord{}, false, errIdempotencyConflict
	}

	return existing, false, nil
}

// complete stores a replayable successful response.
func (store idempotencyStore) complete(ctx context.Context, key string, record idempotencyRecord, response Response) error {
	record.State = "complete"
	record.Response = &response
	payload, err := json.Marshal(record)
	if err != nil {
		return err
	}

	return store.client.Set(ctx, store.key(strings.TrimSpace(key)), payload, idempotencyTTL)
}

// release removes a failed request so it can be retried.
func (store idempotencyStore) release(ctx context.Context, key string) error {
	return store.client.Delete(ctx, store.key(strings.TrimSpace(key)))
}

// find reads one idempotency record.
func (store idempotencyStore) find(ctx context.Context, key string) (idempotencyRecord, bool, error) {
	payload, found, err := store.client.Find(ctx, store.key(key))
	if err != nil || !found {
		return idempotencyRecord{}, found, err
	}

	var record idempotencyRecord
	if err := json.Unmarshal(payload, &record); err != nil {
		return idempotencyRecord{}, false, err
	}

	return record, true, nil
}

// key returns the namespaced Redis key.
func (store idempotencyStore) key(key string) string {
	return idempotencyPrefix + key
}

// requestHash returns a stable request body digest.
func requestHash(request CreateRequest) (string, error) {
	request.Username = strings.TrimSpace(request.Username)
	payload, err := json.Marshal(request)
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256(payload)

	return hex.EncodeToString(digest[:]), nil
}
