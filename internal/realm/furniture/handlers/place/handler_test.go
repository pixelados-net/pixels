package place

import (
	"testing"

	placecmd "github.com/niflaot/pixels/internal/realm/furniture/commands/place"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inplace "github.com/niflaot/pixels/networking/inbound/furniture/place"
)

// TestRegisterAddsPlaceHandler verifies registry wiring.
func TestRegisterAddsPlaceHandler(t *testing.T) {
	registry := netconn.NewHandlerRegistry()
	Register(registry, New(placecmd.Handler{}, nil))

	packet, err := codec.NewPacket(inplace.Header, inplace.Definition, codec.String("1 2 3 0"))
	if err != nil {
		t.Fatalf("new packet: %v", err)
	}
	err = registry.Handle(netconn.Context{State: netconn.StateConnected, Authenticated: true}, packet)
	if err == nil {
		t.Fatal("expected handler dependency error")
	}
}
