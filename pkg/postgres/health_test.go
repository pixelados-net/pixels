package postgres

import (
	"context"
	"errors"
	"testing"
	"time"
)

// TestHealthPingReturnsPingerError verifies health ping failures are returned.
func TestHealthPingReturnsPingerError(t *testing.T) {
	expected := errors.New("offline")
	health := Health{
		Pinger:  &fakePinger{err: expected},
		Timeout: time.Second,
	}

	err := health.Ping(context.Background())
	if !errors.Is(err, expected) {
		t.Fatalf("expected pinger error, got %v", err)
	}
}

// TestHealthPingUsesTimeout verifies health ping receives a deadline.
func TestHealthPingUsesTimeout(t *testing.T) {
	pinger := &fakePinger{}
	health := Health{
		Pinger:  pinger,
		Timeout: time.Second,
	}

	if err := health.Ping(context.Background()); err != nil {
		t.Fatalf("ping health: %v", err)
	}

	if !pinger.deadline {
		t.Fatal("expected ping context deadline")
	}
}

// fakePinger records ping calls for tests.
type fakePinger struct {
	// err is the error returned by Ping.
	err error

	// deadline reports whether Ping received a deadline.
	deadline bool
}

// Ping verifies connectivity for tests.
func (pinger *fakePinger) Ping(ctx context.Context) error {
	_, pinger.deadline = ctx.Deadline()

	return pinger.err
}
