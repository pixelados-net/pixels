package buy

import (
	"testing"

	buycmd "github.com/niflaot/pixels/internal/realm/catalog/commands/buy"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inbuy "github.com/niflaot/pixels/networking/inbound/catalog/item/buy"
	"go.uber.org/zap"
)

// TestRegisterAddsCatalogBuyHandler verifies handler registration.
func TestRegisterAddsCatalogBuyHandler(t *testing.T) {
	registry := netconn.NewHandlerRegistry()
	Register(registry, func(netconn.Context, codec.Packet) error { return nil })
	if registry.Len() != 1 {
		t.Fatalf("expected one handler, got %d", registry.Len())
	}
}

// TestNewDecodesAndDispatchesCatalogBuy verifies packet adapter behavior.
func TestNewDecodesAndDispatchesCatalogBuy(t *testing.T) {
	packet, _ := codec.NewPacket(inbuy.Header, inbuy.Definition, codec.Int32(2), codec.Int32(5), codec.String(""), codec.Int32(1))
	if err := New(buycmd.Handler{}, zap.NewNop())(netconn.Context{}, packet); err == nil {
		t.Fatal("expected unresolved session error after dispatch")
	}
	if err := New(buycmd.Handler{}, zap.NewNop())(netconn.Context{}, codec.Packet{Header: inbuy.Header}); err == nil {
		t.Fatal("expected malformed packet error")
	}
}
