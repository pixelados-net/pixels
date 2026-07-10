package look

import (
	"testing"

	lookcmd "github.com/niflaot/pixels/internal/realm/room/commands/look"
	roomlive "github.com/niflaot/pixels/internal/realm/room/live"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inlook "github.com/niflaot/pixels/networking/inbound/room/entities/look"
)

// TestRegisterAddsLookHandler verifies registry wiring.
func TestRegisterAddsLookHandler(t *testing.T) {
	registry := netconn.NewHandlerRegistry()
	Register(registry, New(lookcmd.Handler{}, nil))

	packet, err := codec.NewPacket(inlook.Header, inlook.Definition, codec.Int32(1), codec.Int32(0))
	if err != nil {
		t.Fatalf("new packet: %v", err)
	}
	err = registry.Handle(netconn.Context{State: netconn.StateConnected, Authenticated: true}, packet)
	if err == nil {
		t.Fatal("expected handler dependency error")
	}
}

// TestStalePresenceErrorsAreSoft verifies post-leave look packets are harmless.
func TestStalePresenceErrorsAreSoft(t *testing.T) {
	if !isStalePresenceError(lookcmd.ErrPlayerNotInRoom) || !isStalePresenceError(roomlive.ErrRoomNotFound) {
		t.Fatal("expected stale room presence errors to be soft")
	}
}
