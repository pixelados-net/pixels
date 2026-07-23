package use

import (
	"testing"

	interactcmd "github.com/niflaot/pixels/internal/realm/furniture/commands/interact"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	incolorwheel "github.com/niflaot/pixels/networking/inbound/furniture/colorwheel"
	indiceactivate "github.com/niflaot/pixels/networking/inbound/furniture/dice/activate"
	indicedeactivate "github.com/niflaot/pixels/networking/inbound/furniture/dice/deactivate"
	inoneway "github.com/niflaot/pixels/networking/inbound/furniture/onewaydoor"
	inrandom "github.com/niflaot/pixels/networking/inbound/furniture/randomstate"
	inuse "github.com/niflaot/pixels/networking/inbound/furniture/use"
	inwall "github.com/niflaot/pixels/networking/inbound/furniture/wallmultistate"
)

// TestRegisterAddsGenericFurnitureUseHandler verifies packet registration and decoding.
func TestRegisterAddsGenericFurnitureUseHandler(t *testing.T) {
	registry := netconn.NewHandlerRegistry()
	Register(registry, New(interactcmd.Handler{}, nil))
	packet, err := codec.NewPacket(inuse.Header, inuse.Definition, codec.Int32(17), codec.Int32(2))
	if err != nil {
		t.Fatalf("create use packet: %v", err)
	}
	if err := registry.Handle(netconn.Context{State: netconn.StateConnected, Authenticated: true}, packet); err == nil {
		t.Fatal("expected missing player dependencies after successful packet routing")
	}
}

// TestRegisterModernBridgesRoutesClientSpecificPackets verifies both regression bridges.
func TestRegisterModernBridgesRoutesClientSpecificPackets(t *testing.T) {
	registry := netconn.NewHandlerRegistry()
	handler := interactcmd.Handler{}
	RegisterModernBridges(registry, NewOneWayDoor(handler, nil), NewRandomState(handler, nil))
	for _, test := range []struct {
		header     uint16
		definition codec.Definition
		values     []codec.Value
	}{{header: inoneway.Header, definition: inoneway.Definition, values: []codec.Value{codec.Int32(17)}},
		{header: inrandom.Header, definition: inrandom.Definition, values: []codec.Value{codec.Int32(17), codec.Int32(2)}}} {
		packet, err := codec.NewPacket(test.header, test.definition, test.values...)
		if err != nil {
			t.Fatal(err)
		}
		if err = registry.Handle(netconn.Context{State: netconn.StateConnected, Authenticated: true}, packet); err == nil {
			t.Fatalf("expected dependencies error for %d", test.header)
		}
	}
}

// TestRegisterDedicatedRoutesNativeInteractionPackets verifies Nitro-specific packet adapters.
func TestRegisterDedicatedRoutesNativeInteractionPackets(t *testing.T) {
	registry := netconn.NewHandlerRegistry()
	handler := interactcmd.Handler{}
	RegisterDedicated(registry, NewDiceActivate(handler, nil), NewDiceClose(handler, nil), NewColorWheel(handler, nil), NewWall(handler, nil))
	tests := []struct {
		header     uint16
		definition codec.Definition
		values     []codec.Value
	}{
		{header: indiceactivate.Header, definition: indiceactivate.Definition, values: []codec.Value{codec.Int32(17)}},
		{header: indicedeactivate.Header, definition: indicedeactivate.Definition, values: []codec.Value{codec.Int32(17)}},
		{header: incolorwheel.Header, definition: incolorwheel.Definition, values: []codec.Value{codec.Int32(17)}},
		{header: inwall.Header, definition: inwall.Definition, values: []codec.Value{codec.Int32(17), codec.Int32(0)}},
	}
	for _, test := range tests {
		packet, err := codec.NewPacket(test.header, test.definition, test.values...)
		if err != nil {
			t.Fatalf("create packet %d: %v", test.header, err)
		}
		if err := registry.Handle(netconn.Context{State: netconn.StateConnected, Authenticated: true}, packet); err == nil {
			t.Fatalf("expected dependencies error for packet %d", test.header)
		}
	}
}
