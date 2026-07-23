package wardrobe

import (
	"context"
	"testing"

	playerconnected "github.com/niflaot/pixels/internal/realm/player/events/connected"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	outclothing "github.com/niflaot/pixels/networking/outbound/user/clothing/sets"
	"github.com/niflaot/pixels/pkg/bus"
	"go.uber.org/fx/fxtest"
)

// TestRegisterBootstrapProjectsClothing verifies connected-player projection and cleanup.
func TestRegisterBootstrapProjectsClothing(t *testing.T) {
	local := bus.New()
	lifecycle := fxtest.NewLifecycle(t)
	connections := netconn.NewRegistry()
	outbound := netconn.NewHandlerRegistry()
	outbound.SetFallback(func(netconn.Context, codec.Packet) error { return nil }, netconn.AllowAnyActiveState(), netconn.AllowUnauthenticated())
	packets := make([]codec.Packet, 0, 1)
	session, _ := netconn.NewSession(netconn.SessionConfig{ID: "wardrobe-bootstrap", Kind: "websocket", Outbound: outbound, Sender: func(_ context.Context, packet codec.Packet) error { packets = append(packets, packet); return nil }, Disposer: func(context.Context, netconn.Reason) error { return nil }})
	_ = connections.Register(session)
	if err := RegisterBootstrap(lifecycle, local, New(&wardrobeStore{}), connections); err != nil {
		t.Fatal(err)
	}
	lifecycle.RequireStart()
	if err := local.Publish(context.Background(), bus.Event{Name: playerconnected.Name, Payload: "invalid"}); err != nil {
		t.Fatal(err)
	}
	if err := local.Publish(context.Background(), bus.Event{Name: playerconnected.Name, Payload: playerconnected.Payload{PlayerID: 7, ConnectionID: "wardrobe-bootstrap", ConnectionKind: "websocket"}}); err != nil {
		t.Fatal(err)
	}
	if len(packets) != 1 || packets[0].Header != outclothing.Header {
		t.Fatalf("packets=%#v", packets)
	}
	lifecycle.RequireStop()
}
