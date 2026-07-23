package websocket

import (
	"context"
	"testing"
	"time"

	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	"go.uber.org/zap"
)

// TestHeartbeatSendsPing verifies active heartbeat pings.
func TestHeartbeatSendsPing(t *testing.T) {
	socket, sent := testSocket(t, time.Now())
	if !socket.heartbeat(context.Background()) {
		t.Fatal("expected active heartbeat")
	}

	if *sent != 1 {
		t.Fatalf("expected one ping, got %d", *sent)
	}
}

// TestHeartbeatDisconnectsIdle verifies idle timeout behavior.
func TestHeartbeatDisconnectsIdle(t *testing.T) {
	socket, _ := testSocket(t, time.Now().Add(-time.Hour))
	if socket.heartbeat(context.Background()) {
		t.Fatal("expected idle disconnect")
	}

	if socket.session.State() != netconn.StateClosed {
		t.Fatalf("expected closed session, got %d", socket.session.State())
	}
}

// testSocket creates a heartbeat socket fixture.
func testSocket(t *testing.T, startedAt time.Time) (*socketSession, *int) {
	t.Helper()
	outbound := netconn.NewHandlerRegistry()
	outbound.SetFallback(func(netconn.Context, codec.Packet) error { return nil }, netconn.AllowAnyActiveState(), netconn.AllowUnauthenticated())
	sent := 0
	session, err := netconn.NewSession(netconn.SessionConfig{
		ID:        "one",
		Kind:      kind,
		StartedAt: startedAt,
		Outbound:  outbound,
		Sender: func(context.Context, codec.Packet) error {
			sent++
			return nil
		},
		Disposer: func(context.Context, netconn.Reason) error {
			return nil
		},
	})
	if err != nil {
		t.Fatalf("new session: %v", err)
	}

	return &socketSession{
		config:  Config{PongTimeout: time.Second}.Normalize(),
		session: session,
		log:     zap.NewNop(),
	}, &sent
}
