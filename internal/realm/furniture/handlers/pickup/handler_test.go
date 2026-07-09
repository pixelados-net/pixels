package pickup

import (
	"testing"

	pickupcmd "github.com/niflaot/pixels/internal/realm/furniture/commands/pickup"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inpickup "github.com/niflaot/pixels/networking/inbound/furniture/pickup"
)

// TestRegisterAddsPickupHandler verifies registry wiring.
func TestRegisterAddsPickupHandler(t *testing.T) {
	registry := netconn.NewHandlerRegistry()
	Register(registry, New(pickupcmd.Handler{}, nil))

	packet, err := codec.NewPacket(inpickup.Header, inpickup.Definition,
		codec.Int32(inpickup.FloorCategory), codec.Int32(1))
	if err != nil {
		t.Fatalf("new packet: %v", err)
	}
	err = registry.Handle(netconn.Context{State: netconn.StateConnected, Authenticated: true}, packet)
	if err == nil {
		t.Fatal("expected handler dependency error")
	}
}
