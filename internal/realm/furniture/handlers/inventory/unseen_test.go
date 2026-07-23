package inventory

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inunseencategory "github.com/niflaot/pixels/networking/inbound/inventory/unseen/category"
	inunseenitems "github.com/niflaot/pixels/networking/inbound/inventory/unseen/items"
)

// TestRegisterUnseenAcceptsNitroAcknowledgements verifies both transient packets are handled.
func TestRegisterUnseenAcceptsNitroAcknowledgements(t *testing.T) {
	registry := netconn.NewHandlerRegistry()
	RegisterUnseen(registry)
	context := netconn.Context{State: netconn.StateConnected, Authenticated: true}
	category, _ := codec.NewPacket(inunseencategory.Header, codec.Definition{codec.Int32Field}, codec.Int32(1))
	items, _ := codec.NewPacket(inunseenitems.Header, codec.Definition{codec.Int32Field, codec.Int32Field, codec.Int32Field}, codec.Int32(1), codec.Int32(1), codec.Int32(7))
	if err := registry.Handle(context, category); err != nil {
		t.Fatalf("handle unseen category: %v", err)
	}
	if err := registry.Handle(context, items); err != nil {
		t.Fatalf("handle unseen items: %v", err)
	}
}
