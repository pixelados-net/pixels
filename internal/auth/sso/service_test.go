package sso

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/niflaot/pixels/pkg/redis"
)

// TestServiceCreateConsumeOnce verifies one-time ticket consumption.
func TestServiceCreateConsumeOnce(t *testing.T) {
	service, server := testService(t)
	ctx := context.Background()

	ticket, err := service.Create(ctx, CreateRequest{PlayerID: 2, IP: "127.0.0.1", TTL: time.Minute})
	if err != nil {
		t.Fatalf("create ticket: %v", err)
	}

	consumed, err := service.Consume(ctx, ConsumeRequest{Ticket: ticket.Value, IP: "127.0.0.1"})
	if err != nil {
		t.Fatalf("consume ticket: %v", err)
	}

	if consumed.PlayerID != 2 {
		t.Fatalf("expected player id, got %d", consumed.PlayerID)
	}

	_, err = service.Consume(ctx, ConsumeRequest{Ticket: ticket.Value, IP: "127.0.0.1"})
	if !errors.Is(err, ErrTicketNotFound) {
		t.Fatalf("expected ticket not found, got %v", err)
	}

	if server.Exists("pixels:sso:" + ticket.Value) {
		t.Fatal("expected opaque ticket not to be a Redis key")
	}
}

// TestServiceConsumeRejectsIPMismatch verifies optional IP binding.
func TestServiceConsumeRejectsIPMismatch(t *testing.T) {
	service, _ := testService(t)
	ctx := context.Background()

	ticket, err := service.Create(ctx, CreateRequest{PlayerID: 2, IP: "127.0.0.1"})
	if err != nil {
		t.Fatalf("create ticket: %v", err)
	}

	_, err = service.Consume(ctx, ConsumeRequest{Ticket: ticket.Value, IP: "10.0.0.1"})
	if !errors.Is(err, ErrTicketIPMismatch) {
		t.Fatalf("expected ip mismatch, got %v", err)
	}
}

// TestServiceCreateRejectsMissingUser verifies user id validation.
func TestServiceCreateRejectsMissingUser(t *testing.T) {
	service, _ := testService(t)
	_, err := service.Create(context.Background(), CreateRequest{})
	if !errors.Is(err, ErrInvalidTicket) {
		t.Fatalf("expected invalid ticket, got %v", err)
	}
}

// TestServiceConsumeRejectsMissingTicket verifies ticket validation.
func TestServiceConsumeRejectsMissingTicket(t *testing.T) {
	service, _ := testService(t)
	_, err := service.Consume(context.Background(), ConsumeRequest{})
	if !errors.Is(err, ErrInvalidTicket) {
		t.Fatalf("expected invalid ticket, got %v", err)
	}
}

// testService creates an SSO service backed by miniredis.
func testService(t *testing.T) (*Service, *miniredis.Miniredis) {
	t.Helper()
	server := miniredis.RunT(t)
	client := redis.New(redis.Config{Address: server.Addr()})
	t.Cleanup(func() {
		if err := client.Close(); err != nil {
			t.Fatalf("close redis: %v", err)
		}
	})

	service := New(Config{DefaultTTL: time.Minute, Key: "test-key", Prefix: "pixels:sso"}, client)
	service.now = func() time.Time {
		return time.Unix(100, 0)
	}

	return service, server
}
