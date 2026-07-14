package action

import (
	"errors"
	"testing"

	actioncmd "github.com/niflaot/pixels/internal/realm/room/world/commands/action"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inaction "github.com/niflaot/pixels/networking/inbound/room/entities/action"
	indance "github.com/niflaot/pixels/networking/inbound/room/entities/dance"
	inposture "github.com/niflaot/pixels/networking/inbound/room/entities/posture"
	insign "github.com/niflaot/pixels/networking/inbound/room/entities/sign"
	"go.uber.org/zap"
)

// TestDecodeActionFamilies verifies grouped handlers preserve packet contracts.
func TestDecodeActionFamilies(t *testing.T) {
	tests := []struct {
		name       string
		kind       actioncmd.Kind
		header     uint16
		definition codec.Definition
	}{
		{name: "dance", kind: actioncmd.KindDance, header: indance.Header, definition: indance.Definition},
		{name: "gesture", kind: actioncmd.KindGesture, header: inaction.Header, definition: inaction.Definition},
		{name: "sign", kind: actioncmd.KindSign, header: insign.Header, definition: insign.Definition},
		{name: "posture", kind: actioncmd.KindPosture, header: inposture.Header, definition: inposture.Definition},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			packet, err := codec.NewPacket(test.header, test.definition, codec.Int32(7))
			if err != nil {
				t.Fatal(err)
			}
			value, err := decode(test.kind, packet)
			if err != nil || value != 7 {
				t.Fatalf("value=%d err=%v", value, err)
			}
		})
	}
	if _, err := decode(actioncmd.Kind(99), codec.Packet{}); !errors.Is(err, codec.ErrUnexpectedHeader) {
		t.Fatalf("expected unsupported action error, got %v", err)
	}
}

// TestRegisterRoutesAllHeaders verifies one handler registration per action packet.
func TestRegisterRoutesAllHeaders(t *testing.T) {
	registry := netconn.NewHandlerRegistry()
	seen := make(map[actioncmd.Kind]int)
	Register(registry, func(kind actioncmd.Kind) netconn.Handler {
		return func(netconn.Context, codec.Packet) error {
			seen[kind]++
			return nil
		}
	})
	if registry.Len() != 4 {
		t.Fatalf("expected four handlers, got %d", registry.Len())
	}
	ctx := netconn.Context{Authenticated: true, State: netconn.StateConnected}
	for _, packet := range []codec.Packet{{Header: indance.Header}, {Header: inaction.Header}, {Header: insign.Header}, {Header: inposture.Header}} {
		if err := registry.Handle(ctx, packet); err != nil {
			t.Fatal(err)
		}
	}
	for _, kind := range []actioncmd.Kind{actioncmd.KindDance, actioncmd.KindGesture, actioncmd.KindSign, actioncmd.KindPosture} {
		if seen[kind] != 1 {
			t.Fatalf("kind %d routed %d times", kind, seen[kind])
		}
	}
}

// TestNewBuildsActionFactories verifies dispatcher construction for every kind.
func TestNewBuildsActionFactories(t *testing.T) {
	factory := New(actioncmd.Handler{}, zap.NewNop())
	for _, kind := range []actioncmd.Kind{actioncmd.KindDance, actioncmd.KindGesture, actioncmd.KindSign, actioncmd.KindPosture} {
		if factory(kind) == nil {
			t.Fatalf("missing factory for kind %d", kind)
		}
	}
	packet, err := codec.NewPacket(indance.Header, indance.Definition, codec.Int32(1))
	if err != nil {
		t.Fatal(err)
	}
	if err = factory(actioncmd.KindDance)(netconn.Context{}, packet); err == nil {
		t.Fatal("expected missing session binding")
	}
}
