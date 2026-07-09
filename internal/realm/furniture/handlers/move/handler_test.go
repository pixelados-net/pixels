package move

import (
	"testing"

	movecmd "github.com/niflaot/pixels/internal/realm/furniture/commands/move"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	infloorupdate "github.com/niflaot/pixels/networking/inbound/furniture/floorupdate"
)

// TestRegisterAddsMoveHandler verifies registry wiring.
func TestRegisterAddsMoveHandler(t *testing.T) {
	registry := netconn.NewHandlerRegistry()
	Register(registry, New(movecmd.Handler{}, nil))

	packet, err := codec.NewPacket(infloorupdate.Header, infloorupdate.Definition,
		codec.Int32(1), codec.Int32(2), codec.Int32(3), codec.Int32(0))
	if err != nil {
		t.Fatalf("new packet: %v", err)
	}
	err = registry.Handle(netconn.Context{State: netconn.StateConnected, Authenticated: true}, packet)
	if err == nil {
		t.Fatal("expected handler dependency error")
	}
}
