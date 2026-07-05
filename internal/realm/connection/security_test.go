package connection

import (
	"context"
	"testing"
	"time"

	"github.com/niflaot/pixels/internal/auth/sso"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	outmachine "github.com/niflaot/pixels/networking/outbound/security/machine"
)

// TestMachineHandlerSendsReplacement verifies invalid machine ids are replaced.
func TestMachineHandlerSendsReplacement(t *testing.T) {
	service := testSSO(t)
	handlers := NewHandlers(service)
	sent := make([]codec.Packet, 0)
	session := testSession(t, handlers, &sent)

	if err := session.Receive(context.Background(), releasePacket(t)); err != nil {
		t.Fatalf("receive release: %v", err)
	}

	if err := session.Receive(context.Background(), machinePacket(t, "~bad")); err != nil {
		t.Fatalf("receive machine: %v", err)
	}

	if len(sent) != 1 || sent[0].Header != outmachine.Header {
		t.Fatalf("expected machine replacement, got %#v", sent)
	}
}

// TestMachineHandlerAcceptsValidMachine verifies accepted machine ids.
func TestMachineHandlerAcceptsValidMachine(t *testing.T) {
	service := testSSO(t)
	handlers := NewHandlers(service)
	sent := make([]codec.Packet, 0)
	session := testSession(t, handlers, &sent)

	if err := session.Receive(context.Background(), releasePacket(t)); err != nil {
		t.Fatalf("receive release: %v", err)
	}
	if err := session.Receive(context.Background(), machinePacket(t, validMachineID())); err != nil {
		t.Fatalf("receive machine: %v", err)
	}
	if len(sent) != 0 {
		t.Fatalf("expected no response, got %#v", sent)
	}
}

// TestTicketHandlerRejectsInvalidTicket verifies failed authentication disposal.
func TestTicketHandlerRejectsInvalidTicket(t *testing.T) {
	service := testSSO(t)
	handlers := NewHandlers(service)
	sent := make([]codec.Packet, 0)
	session := testSession(t, handlers, &sent)

	if err := session.Receive(context.Background(), releasePacket(t)); err != nil {
		t.Fatalf("receive release: %v", err)
	}
	if err := session.Receive(context.Background(), ticketPacket(t, "missing")); err != nil {
		t.Fatalf("receive ticket: %v", err)
	}
	if session.State() != netconn.StateClosed {
		t.Fatalf("expected closed session, got %d", session.State())
	}
}

// TestTicketHandlerRequiresSecurityInProduction verifies encryption policy.
func TestTicketHandlerRequiresSecurityInProduction(t *testing.T) {
	service := testSSO(t)
	ticket, err := service.Create(context.Background(), sso.CreateRequest{UserID: "todo-user", TTL: time.Minute})
	if err != nil {
		t.Fatalf("create ticket: %v", err)
	}

	handlers := NewHandlers(service)
	sent := make([]codec.Packet, 0)
	session := testSession(t, handlers, &sent)
	if err := session.SetSecurityPolicy(netconn.SecurityPolicy{Mode: netconn.SecurityRequired}); err != nil {
		t.Fatalf("set policy: %v", err)
	}

	if err := session.Receive(context.Background(), releasePacket(t)); err != nil {
		t.Fatalf("receive release: %v", err)
	}
	if err := session.Receive(context.Background(), ticketPacket(t, ticket.Value)); err != netconn.ErrSecurityRequired {
		t.Fatalf("expected security required, got %v", err)
	}
	if session.State() != netconn.StateClosed {
		t.Fatalf("expected closed session, got %d", session.State())
	}
}

// machinePacket creates a machine packet.
func machinePacket(t *testing.T, machineID string) codec.Packet {
	t.Helper()
	packet, err := codec.NewPacket(2490, codec.Definition{
		codec.Named("machineId", codec.StringField),
		codec.Named("fingerprint", codec.StringField),
		codec.Named("capabilities", codec.StringField),
	}, codec.String(machineID), codec.String("fingerprint"), codec.String("capabilities"))
	if err != nil {
		t.Fatalf("new machine packet: %v", err)
	}

	return packet
}

// validMachineID returns a syntactically valid test machine id.
func validMachineID() string {
	return "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
}
