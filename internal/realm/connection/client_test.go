package connection

import (
	"context"
	"testing"

	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	outlatency "github.com/niflaot/pixels/networking/outbound/client/latency"
)

// TestConnectedHandlers verifies pong and latency handlers after auth.
func TestConnectedHandlers(t *testing.T) {
	service := testSSO(t)
	handlers := NewHandlers(service)
	sent := make([]codec.Packet, 0)
	session := testSession(t, handlers, &sent)
	mustConnected(t, session)

	if err := session.Receive(context.Background(), latencyPacket(t, 7)); err != nil {
		t.Fatalf("receive latency: %v", err)
	}
	if err := session.Receive(context.Background(), pongPacket(t)); err != nil {
		t.Fatalf("receive pong: %v", err)
	}

	if len(sent) != 1 || sent[0].Header != outlatency.Header {
		t.Fatalf("expected latency response, got %#v", sent)
	}
}

// TestClientHandlersRejectMalformedPayloads verifies decode failures.
func TestClientHandlersRejectMalformedPayloads(t *testing.T) {
	if err := latencyHandler(netconn.Context{}, codec.Packet{Header: 295}); err == nil {
		t.Fatal("expected latency decode failure")
	}
	if err := pongHandler(netconn.Context{}, codec.Packet{Header: 2596, Payload: []byte{1}}); err == nil {
		t.Fatal("expected pong decode failure")
	}
}

// latencyPacket creates a client latency packet.
func latencyPacket(t *testing.T, requestID int32) codec.Packet {
	t.Helper()
	packet, err := codec.NewPacket(295, codec.Definition{codec.Named("requestId", codec.Int32Field)}, codec.Int32(requestID))
	if err != nil {
		t.Fatalf("new latency packet: %v", err)
	}

	return packet
}

// pongPacket creates a client pong packet.
func pongPacket(t *testing.T) codec.Packet {
	t.Helper()
	packet, err := codec.NewPacket(2596, codec.Definition{})
	if err != nil {
		t.Fatalf("new pong packet: %v", err)
	}

	return packet
}
