package heartbeat

import (
	"context"
	"testing"
	"time"

	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inpong "github.com/niflaot/pixels/networking/inbound/client/pong"
)

// TestHandlerMarksPong verifies heartbeat pongs update the session.
func TestHandlerMarksPong(t *testing.T) {
	startedAt := time.Unix(10, 0)
	session := testSession(t, startedAt)

	if err := session.Receive(context.Background(), pongPacket(t)); err != nil {
		t.Fatalf("receive pong: %v", err)
	}
	if !session.LastPongAt().After(startedAt) {
		t.Fatalf("expected updated pong time, got %s", session.LastPongAt())
	}
}

// TestHandlerRejectsMalformedPayload verifies decode failures.
func TestHandlerRejectsMalformedPayload(t *testing.T) {
	err := Handler(netconn.Context{}, codec.Packet{Header: inpong.Header, Payload: []byte{1}})
	if err == nil {
		t.Fatal("expected pong decode failure")
	}
}

// testSession creates a connected heartbeat session.
func testSession(t *testing.T, startedAt time.Time) *netconn.Session {
	t.Helper()
	inbound := netconn.NewHandlerRegistry()
	Register(inbound)
	session, err := netconn.NewSession(netconn.SessionConfig{
		ID:        "heartbeat-test",
		Kind:      "websocket",
		StartedAt: startedAt,
		Inbound:   inbound,
		Sender: func(context.Context, codec.Packet) error {
			return nil
		},
		Disposer: func(context.Context, netconn.Reason) error {
			return nil
		},
	})
	if err != nil {
		t.Fatalf("new session: %v", err)
	}

	mustConnected(t, session)

	return session
}

// mustConnected moves a session to connected state.
func mustConnected(t *testing.T, session *netconn.Session) {
	t.Helper()
	if err := session.Transition(netconn.EventPacketReceived); err != nil {
		t.Fatalf("packet transition: %v", err)
	}
	if err := session.Transition(netconn.EventAuthenticationStarted); err != nil {
		t.Fatalf("auth transition: %v", err)
	}
	if err := session.Authenticate(time.Now()); err != nil {
		t.Fatalf("authenticate: %v", err)
	}
	if err := session.Transition(netconn.EventSessionReady); err != nil {
		t.Fatalf("ready transition: %v", err)
	}
}

// pongPacket creates a client pong packet.
func pongPacket(t *testing.T) codec.Packet {
	t.Helper()
	packet, err := codec.NewPacket(inpong.Header, inpong.Definition)
	if err != nil {
		t.Fatalf("new pong packet: %v", err)
	}

	return packet
}
