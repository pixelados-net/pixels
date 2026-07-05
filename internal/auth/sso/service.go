package sso

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/niflaot/pixels/pkg/redis"
)

var (
	// ErrInvalidTicket reports a malformed ticket.
	ErrInvalidTicket = errors.New("invalid sso ticket")

	// ErrTicketIPMismatch reports a ticket consumed from the wrong IP.
	ErrTicketIPMismatch = errors.New("sso ticket ip mismatch")

	// ErrTicketNotFound reports a missing, expired, or already consumed ticket.
	ErrTicketNotFound = errors.New("sso ticket not found")
)

// Service creates and consumes one-time SSO tickets.
type Service struct {
	config Config
	redis  *redis.Client
	now    func() time.Time
}

// New creates an SSO service.
func New(config Config, redis *redis.Client) *Service {
	return &Service{
		config: config,
		redis:  redis,
		now:    time.Now,
	}
}

// Create creates and stores a one-time SSO ticket.
func (service *Service) Create(ctx context.Context, request CreateRequest) (Ticket, error) {
	if strings.TrimSpace(request.UserID) == "" {
		return Ticket{}, ErrInvalidTicket
	}

	ttl := service.ttl(request.TTL)
	value, err := randomTicket()
	if err != nil {
		return Ticket{}, err
	}

	ticket := Ticket{
		Value:     value,
		UserID:    request.UserID,
		IP:        request.IP,
		ExpiresAt: service.now().Add(ttl),
	}
	payload, err := json.Marshal(record{UserID: ticket.UserID, IP: ticket.IP, ExpiresAt: ticket.ExpiresAt})
	if err != nil {
		return Ticket{}, err
	}

	if err := service.redis.Set(ctx, service.key(value), payload, ttl); err != nil {
		return Ticket{}, err
	}

	return ticket, nil
}

// Consume consumes a one-time SSO ticket.
func (service *Service) Consume(ctx context.Context, request ConsumeRequest) (Ticket, error) {
	if strings.TrimSpace(request.Ticket) == "" {
		return Ticket{}, ErrInvalidTicket
	}

	payload, ok, err := service.redis.Take(ctx, service.key(request.Ticket))
	if err != nil {
		return Ticket{}, err
	}

	if !ok {
		return Ticket{}, ErrTicketNotFound
	}

	var data record
	if err := json.Unmarshal(payload, &data); err != nil {
		return Ticket{}, err
	}

	if data.IP != "" && data.IP != request.IP {
		return Ticket{}, ErrTicketIPMismatch
	}

	return Ticket{Value: request.Ticket, UserID: data.UserID, IP: data.IP, ExpiresAt: data.ExpiresAt}, nil
}

// ttl returns the request TTL or default TTL.
func (service *Service) ttl(ttl time.Duration) time.Duration {
	if ttl > 0 {
		return ttl
	}

	return service.config.DefaultTTL
}

// key returns the Redis storage key for a ticket.
func (service *Service) key(ticket string) string {
	mac := hmac.New(sha256.New, []byte(service.config.Key))
	_, _ = mac.Write([]byte(ticket))

	return service.config.Prefix + ":" + hex.EncodeToString(mac.Sum(nil))
}

// randomTicket returns a base64url random ticket.
func randomTicket() (string, error) {
	value := make([]byte, 32)
	if _, err := rand.Read(value); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(value), nil
}
