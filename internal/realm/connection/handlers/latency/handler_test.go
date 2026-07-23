package latency

import (
	"context"
	"testing"
	"time"

	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inlatency "github.com/niflaot/pixels/networking/inbound/client/latency"
	outlatency "github.com/niflaot/pixels/networking/outbound/client/latency"
)

// TestHandlerSendsLatency verifies latency echo responses.
func TestHandlerSendsLatency(t *testing.T) {
	sent := make([]codec.Packet, 0)
	session := testSession(t, &sent)

	if err := session.Receive(context.Background(), latencyPacket(t, 7)); err != nil {
		t.Fatalf("receive latency: %v", err)
	}
	if len(sent) != 1 || sent[0].Header != outlatency.Header {
		t.Fatalf("expected latency response, got %#v", sent)
	}
}

// TestHandlerRejectsMalformedPayload verifies decode failures.
func TestHandlerRejectsMalformedPayload(t *testing.T) {
	err := Handler(netconn.Context{}, codec.Packet{Header: inlatency.Header})
	if err == nil {
		t.Fatal("expected latency decode failure")
	}
}

// testSession creates a connected latency session.
func testSession(t *testing.T, sent *[]codec.Packet) *netconn.Session {
	t.Helper()
	inbound := netconn.NewHandlerRegistry()
	Register(inbound)
	outbound := netconn.NewHandlerRegistry()
	outbound.SetFallback(func(netconn.Context, codec.Packet) error {
		return nil
	}, netconn.AllowAnyActiveState(), netconn.AllowUnauthenticated())
	session, err := netconn.NewSession(netconn.SessionConfig{
		ID:       "latency-test",
		Kind:     "websocket",
		Inbound:  inbound,
		Outbound: outbound,
		Sender: func(ctx context.Context, packet codec.Packet) error {
			*sent = append(*sent, packet)
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

// latencyPacket creates a client latency packet.
func latencyPacket(t *testing.T, requestID int32) codec.Packet {
	t.Helper()
	packet, err := codec.NewPacket(inlatency.Header, inlatency.Definition, codec.Int32(requestID))
	if err != nil {
		t.Fatalf("new latency packet: %v", err)
	}

	return packet
}
