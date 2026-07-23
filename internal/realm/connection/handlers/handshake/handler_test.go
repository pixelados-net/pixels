package handshake

import (
	"context"
	"testing"

	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	indiffiecomplete "github.com/niflaot/pixels/networking/inbound/handshake/diffie/complete"
	indiffieinit "github.com/niflaot/pixels/networking/inbound/handshake/diffie/init"
	inpolicy "github.com/niflaot/pixels/networking/inbound/handshake/policy"
	inrelease "github.com/niflaot/pixels/networking/inbound/handshake/release"
	invariables "github.com/niflaot/pixels/networking/inbound/handshake/variables"
)

// TestRegisterHandlesEarlyHandshake verifies early metadata packets.
func TestRegisterHandlesEarlyHandshake(t *testing.T) {
	session := testSession(t)

	if err := session.Receive(context.Background(), releasePacket(t)); err != nil {
		t.Fatalf("receive release: %v", err)
	}
	if err := session.Receive(context.Background(), variablesPacket(t)); err != nil {
		t.Fatalf("receive variables: %v", err)
	}
	if err := session.Receive(context.Background(), policyPacket(t)); err != nil {
		t.Fatalf("receive policy: %v", err)
	}
}

// TestDiffieInitUnavailableDisconnects verifies placeholder Diffie failure.
func TestDiffieInitUnavailableDisconnects(t *testing.T) {
	session := testSession(t)

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

// TestDiffieCompleteUnavailableDisconnects verifies completion failure.
func TestDiffieCompleteUnavailableDisconnects(t *testing.T) {
	session := testSession(t)

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

// testSession creates a handshake session.
func testSession(t *testing.T) *netconn.Session {
	t.Helper()
	inbound := netconn.NewHandlerRegistry()
	Register(inbound)
	session, err := netconn.NewSession(netconn.SessionConfig{
		ID:      "handshake-test",
		Kind:    "websocket",
		Inbound: inbound,
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

	return session
}

// releasePacket creates a release-version packet.
func releasePacket(t *testing.T) codec.Packet {
	t.Helper()
	packet, err := codec.NewPacket(
		inrelease.Header,
		inrelease.Definition,
		codec.String("NITRO-test"),
		codec.String("HTML5"),
		codec.Int32(0),
		codec.Int32(0),
	)
	if err != nil {
		t.Fatalf("new release packet: %v", err)
	}

	return packet
}

// variablesPacket creates a variables packet.
func variablesPacket(t *testing.T) codec.Packet {
	t.Helper()
	packet, err := codec.NewPacket(invariables.Header, invariables.Definition, codec.Int32(1), codec.String("client"), codec.String("vars"))
	if err != nil {
		t.Fatalf("new variables packet: %v", err)
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

// diffieInitPacket creates a Diffie init packet.
func diffieInitPacket(t *testing.T) codec.Packet {
	t.Helper()
	packet, err := codec.NewPacket(indiffieinit.Header, indiffieinit.Definition)
	if err != nil {
		t.Fatalf("new diffie init packet: %v", err)
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
