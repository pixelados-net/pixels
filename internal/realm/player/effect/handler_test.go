package effect

import (
	"context"
	"testing"

	playerconnected "github.com/niflaot/pixels/internal/realm/player/events/connected"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inactivate "github.com/niflaot/pixels/networking/inbound/user/effect/activate"
	inenable "github.com/niflaot/pixels/networking/inbound/user/effect/enable"
	"github.com/niflaot/pixels/pkg/bus"
	"go.uber.org/fx/fxtest"
	"go.uber.org/zap"
)

// TestEffectHandlersActivateAndSelect verifies authenticated packet adapters.
func TestEffectHandlersActivateAndSelect(t *testing.T) {
	store := newMemoryStore()
	service := New(store, nil, nil, nil, nil, nil)
	if _, err := service.Grant(context.Background(), 7, 101, 60, SourceAdmin); err != nil {
		t.Fatal(err)
	}
	bindings := binding.NewRegistry()
	if err := bindings.Add(binding.Binding{PlayerID: 7, ConnectionID: "effect", ConnectionKind: "websocket"}); err != nil {
		t.Fatal(err)
	}
	handler := Handler{Effects: service, Bindings: bindings, Log: zap.NewNop()}
	registry := netconn.NewHandlerRegistry()
	RegisterHandlers(registry, handler)
	ctx := netconn.Context{ConnectionID: "effect", ConnectionKind: "websocket", Authenticated: true, State: netconn.StateConnected}
	enablePacket, err := codec.NewPacket(inenable.Header, inenable.Definition, codec.Int32(101))
	if err != nil {
		t.Fatal(err)
	}
	if err = registry.Handle(ctx, enablePacket); err != nil {
		t.Fatal(err)
	}
	activatePacket, err := codec.NewPacket(inactivate.Header, inactivate.Definition, codec.Int32(101))
	if err != nil {
		t.Fatal(err)
	}
	if err = registry.Handle(ctx, activatePacket); err != nil {
		t.Fatal(err)
	}
	if store.active[7] == nil || *store.active[7] != 101 || store.effects[7][101].ActivatedAt == nil {
		t.Fatalf("unexpected effect state active=%v effect=%#v", store.active[7], store.effects[7][101])
	}
}

// TestEffectHandlersIgnoreUnboundAndNilServices verifies stale packet handling.
func TestEffectHandlersIgnoreUnboundAndNilServices(t *testing.T) {
	handler := Handler{}
	packet, err := codec.NewPacket(inenable.Header, inenable.Definition, codec.Int32(101))
	if err != nil {
		t.Fatal(err)
	}
	if err = handler.enable(netconn.Context{}, packet); err != nil {
		t.Fatal(err)
	}
	activatePacket, err := codec.NewPacket(inactivate.Header, inactivate.Definition, codec.Int32(101))
	if err != nil {
		t.Fatal(err)
	}
	if err = handler.activate(netconn.Context{}, activatePacket); err != nil {
		t.Fatal(err)
	}
	RegisterHandlers(nil, handler)
	if _, found := handler.playerID(netconn.Context{}); found {
		t.Fatal("unexpected player binding")
	}
}

// TestRegisterBootstrapProjectsConnectedPlayers verifies lifecycle subscription wiring.
func TestRegisterBootstrapProjectsConnectedPlayers(t *testing.T) {
	local := bus.New()
	lifecycle := fxtest.NewLifecycle(t)
	service := New(newMemoryStore(), nil, nil, nil, nil, nil)
	if err := RegisterBootstrap(lifecycle, local, service); err != nil {
		t.Fatal(err)
	}
	lifecycle.RequireStart()
	if err := local.Publish(context.Background(), bus.Event{Name: playerconnected.Name, Payload: "invalid"}); err != nil {
		t.Fatal(err)
	}
	if err := local.Publish(context.Background(), bus.Event{Name: playerconnected.Name, Payload: playerconnected.Payload{PlayerID: 7}}); err != nil {
		t.Fatal(err)
	}
	lifecycle.RequireStop()
}
