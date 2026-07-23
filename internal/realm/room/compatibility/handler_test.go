package compatibility

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inclothing "github.com/niflaot/pixels/networking/inbound/room/compatibility/clothing"
	innetwork "github.com/niflaot/pixels/networking/inbound/room/compatibility/network"
	inpromoted "github.com/niflaot/pixels/networking/inbound/room/compatibility/promoted"
	inqueue "github.com/niflaot/pixels/networking/inbound/room/compatibility/queue"
	invote "github.com/niflaot/pixels/networking/inbound/room/compatibility/vote"
)

// TestRegisterDecodesEveryRetiredRoomPacket verifies all NOOP aliases keep the session alive.
func TestRegisterDecodesEveryRetiredRoomPacket(t *testing.T) {
	registry := netconn.NewHandlerRegistry()
	Register(registry)
	tests := []struct {
		header     uint16
		definition codec.Definition
		values     []codec.Value
	}{
		{header: inpromoted.Header, definition: inpromoted.Definition, values: []codec.Value{codec.String("official")}},
		{header: inclothing.Header, definition: inclothing.Definition, values: []codec.Value{codec.Int32(1), codec.String("M"), codec.String("hd-1")}},
		{header: inqueue.Header, definition: inqueue.Definition, values: []codec.Value{codec.Int32(2)}},
		{header: innetwork.Header, definition: innetwork.Definition, values: []codec.Value{codec.Int32(1), codec.Int32(2)}},
		{header: invote.Header, definition: invote.Definition},
	}
	for _, test := range tests {
		packet, err := codec.NewPacket(test.header, test.definition, test.values...)
		if err != nil {
			t.Fatal(err)
		}
		if err = registry.Handle(netconn.Context{State: netconn.StateConnected, Authenticated: true}, packet); err != nil {
			t.Fatalf("header=%d err=%v", test.header, err)
		}
	}
}

// TestHandleRejectsUnknownHeader verifies the compatibility adapter stays strict.
func TestHandleRejectsUnknownHeader(t *testing.T) {
	Register(nil)
	if err := handle(netconn.Context{}, codec.Packet{Header: 65535}); err == nil {
		t.Fatal("expected unexpected header")
	}
}
