package security

import (
	"context"
	"testing"
	"time"

	"github.com/niflaot/pixels/internal/auth/sso"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	outbanned "github.com/niflaot/pixels/networking/outbound/user/session/banned"
)

// bannedGate rejects every player for focused authentication tests.
type bannedGate struct{}

// CheckBan reports one active test ban.
func (bannedGate) CheckBan(context.Context, int64) (bool, string, error) {
	return true, "internal moderation reason", nil
}

// TestTicketProjectsBanBeforeDisconnect verifies Nitro receives USER_BANNED.
func TestTicketProjectsBanBeforeDisconnect(t *testing.T) {
	service := testSSO(t)
	ticket, err := service.Create(context.Background(), sso.CreateRequest{PlayerID: 2, TTL: time.Minute})
	if err != nil {
		t.Fatalf("create ticket: %v", err)
	}
	authenticator := testAuthenticator(t, service)
	authenticator.SetSanctionGate(bannedGate{})
	session, sent := bannedSession(t, authenticator)
	if err = session.Receive(context.Background(), ticketPacket(t, ticket.Value)); err != nil {
		t.Fatalf("receive ticket: %v", err)
	}
	if session.State() != netconn.StateClosed || len(*sent) != 1 || (*sent)[0].Header != outbanned.Header {
		t.Fatalf("state=%d sent=%#v", session.State(), *sent)
	}
}

// bannedSession creates a security session with an explicit authenticator.
func bannedSession(t *testing.T, authenticator *Authenticator) (*netconn.Session, *[]codec.Packet) {
	t.Helper()
	inbound := netconn.NewHandlerRegistry()
	Register(inbound, authenticator)
	outbound := netconn.NewHandlerRegistry()
	outbound.SetFallback(func(netconn.Context, codec.Packet) error { return nil }, netconn.AllowAnyActiveState(), netconn.AllowUnauthenticated())
	sent := make([]codec.Packet, 0, 1)
	session, err := netconn.NewSession(netconn.SessionConfig{ID: "ban-test", Kind: "websocket", Inbound: inbound, Outbound: outbound,
		Sender:   func(_ context.Context, packet codec.Packet) error { sent = append(sent, packet); return nil },
		Disposer: func(context.Context, netconn.Reason) error { return nil }})
	if err != nil {
		t.Fatalf("new session: %v", err)
	}
	if err = session.Transition(netconn.EventPacketReceived); err != nil {
		t.Fatalf("packet transition: %v", err)
	}
	return session, &sent
}
