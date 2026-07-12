package use

import (
	"testing"

	interactcmd "github.com/niflaot/pixels/internal/realm/furniture/commands/interact"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	incolorwheel "github.com/niflaot/pixels/networking/inbound/furniture/colorwheel"
	indiceactivate "github.com/niflaot/pixels/networking/inbound/furniture/dice/activate"
	indicedeactivate "github.com/niflaot/pixels/networking/inbound/furniture/dice/deactivate"
	inuse "github.com/niflaot/pixels/networking/inbound/furniture/use"
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

// TestRegisterDedicatedRoutesNativeInteractionPackets verifies Nitro-specific packet adapters.
func TestRegisterDedicatedRoutesNativeInteractionPackets(t *testing.T) {
	registry := netconn.NewHandlerRegistry()
	handler := interactcmd.Handler{}
	RegisterDedicated(registry, NewDiceActivate(handler, nil), NewDiceClose(handler, nil), NewColorWheel(handler, nil))
	tests := []struct {
		header     uint16
		definition codec.Definition
	}{
		{header: indiceactivate.Header, definition: indiceactivate.Definition},
		{header: indicedeactivate.Header, definition: indicedeactivate.Definition},
		{header: incolorwheel.Header, definition: incolorwheel.Definition},
	}
	for _, test := range tests {
		packet, err := codec.NewPacket(test.header, test.definition, codec.Int32(17))
		if err != nil {
			t.Fatalf("create packet %d: %v", test.header, err)
		}
		if err := registry.Handle(netconn.Context{State: netconn.StateConnected, Authenticated: true}, packet); err == nil {
			t.Fatalf("expected dependencies error for packet %d", test.header)
		}
	}
}
