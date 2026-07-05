package connection

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/niflaot/pixels/internal/auth/sso"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inrelease "github.com/niflaot/pixels/networking/inbound/handshake/release"
	inticket "github.com/niflaot/pixels/networking/inbound/security/ticket"
	outauth "github.com/niflaot/pixels/networking/outbound/authentication/ok"
	"github.com/niflaot/pixels/pkg/redis"
)

// TestHandlersAuthenticateWithSSO verifies the development authentication path.
func TestHandlersAuthenticateWithSSO(t *testing.T) {
	service := testSSO(t)
	ticket, err := service.Create(context.Background(), sso.CreateRequest{UserID: "todo-user", TTL: time.Minute})
	if err != nil {
		t.Fatalf("create ticket: %v", err)
	}

	handlers := NewHandlers(service)
	sent := make([]codec.Packet, 0)
	session, err := netconn.NewSession(netconn.SessionConfig{
		ID:       "one",
		Kind:     "websocket",
		Inbound:  handlers.Inbound,
		Outbound: handlers.Outbound,
		Sender: func(ctx context.Context, packet codec.Packet) error {
			sent = append(sent, packet)
			return nil
		},
		Disposer: func(context.Context, netconn.Reason) error {
			return nil
		},
	})
	if err != nil {
		t.Fatalf("new session: %v", err)
	}

	if err := session.Receive(context.Background(), releasePacket(t)); err != nil {
		t.Fatalf("receive release: %v", err)
	}

	if err := session.Receive(context.Background(), ticketPacket(t, ticket.Value)); err != nil {
		t.Fatalf("receive ticket: %v", err)
	}

	if session.State() != netconn.StateConnected {
		t.Fatalf("expected connected state, got %d", session.State())
	}

	if len(sent) == 0 || sent[0].Header != outauth.Header {
		t.Fatalf("expected authenticated packet, got %#v", sent)
	}
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

// ticketPacket creates a security-ticket packet.
func ticketPacket(t *testing.T, ticket string) codec.Packet {
	t.Helper()
	packet, err := codec.NewPacket(inticket.Header, inticket.Definition, codec.String(ticket), codec.Int32(1))
	if err != nil {
		t.Fatalf("new ticket packet: %v", err)
	}

	return packet
}

// testSSO creates a test SSO service.
func testSSO(t *testing.T) *sso.Service {
	t.Helper()
	server := miniredis.RunT(t)
	client := redis.New(redis.Config{Address: server.Addr()})
	t.Cleanup(func() {
		if err := client.Close(); err != nil {
			t.Fatalf("close redis: %v", err)
		}
	})

	return sso.New(sso.Config{DefaultTTL: time.Minute, Key: "test-key", Prefix: "pixels:sso"}, client)
}

// testSession creates a connection-realm session.
func testSession(t *testing.T, handlers *Handlers, sent *[]codec.Packet) *netconn.Session {
	t.Helper()
	session, err := netconn.NewSession(netconn.SessionConfig{
		ID:       "one",
		Kind:     "websocket",
		Inbound:  handlers.Inbound,
		Outbound: handlers.Outbound,
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
