package connection

import (
	"context"
	"testing"

	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	indiffiecomplete "github.com/niflaot/pixels/networking/inbound/handshake/diffie/complete"
	inpolicy "github.com/niflaot/pixels/networking/inbound/handshake/policy"
	invariables "github.com/niflaot/pixels/networking/inbound/handshake/variables"
)

// TestDiffieUnavailableDisconnects verifies placeholder Diffie failure behavior.
func TestDiffieUnavailableDisconnects(t *testing.T) {
	service := testSSO(t)
	handlers := NewHandlers(service)
	sent := make([]codec.Packet, 0)
	session := testSession(t, handlers, &sent)

	if err := session.Receive(context.Background(), releasePacket(t)); err != nil {
		t.Fatalf("receive release: %v", err)
	}

	if err := session.Receive(context.Background(), diffieInitPacket(t)); err != nil {
		t.Fatalf("receive diffie init: %v", err)
	}

	if session.State() != netconn.StateClosed {
		t.Fatalf("expected closed session, got %d", session.State())
	}
}

// TestEarlyHandshakeHandlers verifies metadata and policy packets.
func TestEarlyHandshakeHandlers(t *testing.T) {
	service := testSSO(t)
	handlers := NewHandlers(service)
	sent := make([]codec.Packet, 0)
	session := testSession(t, handlers, &sent)

	if err := session.Receive(context.Background(), variablesPacket(t)); err != nil {
		t.Fatalf("receive variables: %v", err)
	}
	if err := session.Receive(context.Background(), policyPacket(t)); err != nil {
		t.Fatalf("receive policy: %v", err)
	}
}

// TestDiffieCompleteUnavailableDisconnects verifies completion failure behavior.
func TestDiffieCompleteUnavailableDisconnects(t *testing.T) {
	service := testSSO(t)
	handlers := NewHandlers(service)
	sent := make([]codec.Packet, 0)
	session := testSession(t, handlers, &sent)

	if err := session.Receive(context.Background(), releasePacket(t)); err != nil {
		t.Fatalf("receive release: %v", err)
	}
	if err := session.Transition(netconn.EventDiffieRequested); err != nil {
		t.Fatalf("transition securing: %v", err)
	}
	if err := session.Receive(context.Background(), diffieCompletePacket(t)); err != nil {
		t.Fatalf("receive diffie complete: %v", err)
	}
	if session.State() != netconn.StateClosed {
		t.Fatalf("expected closed session, got %d", session.State())
	}
}

// diffieInitPacket creates a Diffie init packet.
func diffieInitPacket(t *testing.T) codec.Packet {
	t.Helper()
	packet, err := codec.NewPacket(3110, codec.Definition{})
	if err != nil {
		t.Fatalf("new diffie packet: %v", err)
	}

	return packet
}

// diffieCompletePacket creates a Diffie complete packet.
func diffieCompletePacket(t *testing.T) codec.Packet {
	t.Helper()
	packet, err := codec.NewPacket(indiffiecomplete.Header, indiffiecomplete.Definition, codec.String("key"))
	if err != nil {
		t.Fatalf("new diffie complete packet: %v", err)
	}

	return packet
}

// policyPacket creates a policy packet.
func policyPacket(t *testing.T) codec.Packet {
	t.Helper()
	packet, err := codec.NewPacket(inpolicy.Header, inpolicy.Definition)
	if err != nil {
		t.Fatalf("new policy packet: %v", err)
	}

	return packet
}

// variablesPacket creates a variables packet.
func variablesPacket(t *testing.T) codec.Packet {
	t.Helper()
	packet, err := codec.NewPacket(
		invariables.Header,
		invariables.Definition,
		codec.Int32(1),
		codec.String("http://client"),
		codec.String("http://vars"),
	)
	if err != nil {
		t.Fatalf("new variables packet: %v", err)
	}

	return packet
}
