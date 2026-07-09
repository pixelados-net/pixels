package inventory

import (
	"testing"

	inventorycmd "github.com/niflaot/pixels/internal/realm/furniture/commands/inventory"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	ininventory "github.com/niflaot/pixels/networking/inbound/inventory/furniture"
)

// TestRegisterAddsInventoryHandler verifies registry wiring.
func TestRegisterAddsInventoryHandler(t *testing.T) {
	registry := netconn.NewHandlerRegistry()
	Register(registry, New(inventorycmd.Handler{}, nil))

	packet, err := codec.NewPacket(ininventory.Header, ininventory.Definition)
	if err != nil {
		t.Fatalf("new packet: %v", err)
	}
	err = registry.Handle(netconn.Context{State: netconn.StateConnected, Authenticated: true}, packet)
	if err == nil {
		t.Fatal("expected handler dependency error")
	}
}
