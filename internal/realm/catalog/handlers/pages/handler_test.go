package pages

import (
	"testing"

	pagescmd "github.com/niflaot/pixels/internal/realm/catalog/commands/pages"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inpages "github.com/niflaot/pixels/networking/inbound/catalog/mode/request"
	"go.uber.org/zap"
)

// TestRegisterAddsCatalogPagesHandler verifies handler registration.
func TestRegisterAddsCatalogPagesHandler(t *testing.T) {
	registry := netconn.NewHandlerRegistry()
	Register(registry, func(netconn.Context, codec.Packet) error { return nil })
	if registry.Len() != 1 {
		t.Fatalf("expected one handler, got %d", registry.Len())
	}
}

// TestNewDecodesAndDispatchesCatalogPages verifies packet adapter behavior.
func TestNewDecodesAndDispatchesCatalogPages(t *testing.T) {
	packet, _ := codec.NewPacket(inpages.Header, inpages.Definition, codec.String("NORMAL"))
	if err := New(pagescmd.Handler{}, zap.NewNop())(netconn.Context{}, packet); err == nil {
		t.Fatal("expected unresolved session error after dispatch")
	}
	if err := New(pagescmd.Handler{}, zap.NewNop())(netconn.Context{}, codec.Packet{Header: inpages.Header}); err == nil {
		t.Fatal("expected malformed packet error")
	}
}
