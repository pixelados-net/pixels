package control

import (
	"context"
	"errors"
	"testing"

	roomcommand "github.com/niflaot/pixels/internal/realm/room/commands/control"
	roommoderation "github.com/niflaot/pixels/internal/realm/room/moderation"
	roomrights "github.com/niflaot/pixels/internal/realm/room/rights"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	"github.com/niflaot/pixels/pkg/i18n"
	"go.uber.org/zap"
)

// TestTranslationKeyClassifiesOnlyDomainRejections verifies soft-error boundaries.
func TestTranslationKeyClassifiesOnlyDomainRejections(t *testing.T) {
	tests := []struct {
		// name identifies the case.
		name string
		// err stores the tested error.
		err error
		// soft stores the expected classification.
		soft bool
	}{
		{name: "rights denied", err: roomrights.ErrAccessDenied, soft: true},
		{name: "protected", err: roommoderation.ErrTargetProtected, soft: true},
		{name: "owner", err: roommoderation.ErrTargetOwner, soft: true},
		{name: "self", err: roommoderation.ErrSelfTarget, soft: true},
		{name: "invalid", err: roommoderation.ErrInvalidMuteDuration, soft: true},
		{name: "target absent", err: roomcommand.ErrTargetNotInRoom, soft: true},
		{name: "unexpected", err: errors.New("database unavailable"), soft: false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, soft := translationKey(test.err)
			if soft != test.soft {
				t.Fatalf("expected soft=%v, got %v", test.soft, soft)
			}
		})
	}
}

// TestWrapSendsLocalizedBubbleWithoutDisconnecting verifies soft rejection delivery.
func TestWrapSendsLocalizedBubbleWithoutDisconnecting(t *testing.T) {
	inbound := netconn.NewHandlerRegistry()
	outbound := netconn.NewHandlerRegistry()
	outbound.SetFallback(func(netconn.Context, codec.Packet) error { return nil }, netconn.AllowAnyActiveState(), netconn.AllowUnauthenticated())
	translations := i18n.NewCatalog(i18n.Config{DefaultLocale: "en"}, map[i18n.Locale]map[i18n.Key]string{"en": {"session.bubble.room.control.denied": "Denied"}})
	wrapped := Wrap(func(netconn.Context, codec.Packet) error { return roomrights.ErrAccessDenied }, translations, zap.NewNop())
	if err := inbound.Register(1, wrapped, netconn.AllowAnyActiveState(), netconn.AllowUnauthenticated()); err != nil {
		t.Fatalf("register inbound: %v", err)
	}
	sent := make([]codec.Packet, 0, 1)
	session, err := netconn.NewSession(netconn.SessionConfig{
		ID: "conn", Kind: "websocket", Inbound: inbound, Outbound: outbound,
		Sender:   func(_ context.Context, packet codec.Packet) error { sent = append(sent, packet); return nil },
		Disposer: func(context.Context, netconn.Reason) error { return nil },
	})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	if err := session.Receive(context.Background(), codec.Packet{Header: 1}); err != nil {
		t.Fatalf("receive soft rejection: %v", err)
	}
	if len(sent) != 1 || sent[0].Header != 1992 {
		t.Fatalf("expected bubble alert, got %#v", sent)
	}
}
