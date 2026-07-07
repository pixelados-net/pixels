package entrytile

import (
	"testing"

	entrytilecmd "github.com/niflaot/pixels/internal/realm/room/commands/entrytile"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inentrytile "github.com/niflaot/pixels/networking/inbound/room/entrytile"
)

// TestRegisterAddsEntryTileHandler verifies registry wiring.
func TestRegisterAddsEntryTileHandler(t *testing.T) {
	registry := netconn.NewHandlerRegistry()
	Register(registry, New(entrytilecmd.Handler{}, nil))

	err := registry.Handle(netconn.Context{State: netconn.StateConnected, Authenticated: true}, codec.Packet{Header: inentrytile.Header})
	if err == nil {
		t.Fatal("expected handler dependency error")
	}
}
