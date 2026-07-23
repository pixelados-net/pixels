package present

import (
	"testing"

	presentcmd "github.com/niflaot/pixels/internal/realm/furniture/commands/present"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inopen "github.com/niflaot/pixels/networking/inbound/furniture/present/open"
)

// TestRegisterAddsOpenPresentHandler verifies registry wiring.
func TestRegisterAddsOpenPresentHandler(t *testing.T) {
	registry := netconn.NewHandlerRegistry()
	Register(registry, New(presentcmd.Handler{}, nil))

	packet, err := codec.NewPacket(inopen.Header, inopen.Definition, codec.Int32(41))
	if err != nil {
		t.Fatalf("new packet: %v", err)
	}
	if err := registry.Handle(netconn.Context{State: netconn.StateConnected, Authenticated: true}, packet); err == nil {
		t.Fatal("expected missing binding error from registered handler")
	}
}
