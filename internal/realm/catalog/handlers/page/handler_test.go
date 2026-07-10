package page

import (
	"testing"

	pagecmd "github.com/niflaot/pixels/internal/realm/catalog/commands/page"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inpage "github.com/niflaot/pixels/networking/inbound/catalog/page/request"
	"go.uber.org/zap"
)

// TestRegisterAddsCatalogPageHandler verifies handler registration.
func TestRegisterAddsCatalogPageHandler(t *testing.T) {
	registry := netconn.NewHandlerRegistry()
	Register(registry, func(netconn.Context, codec.Packet) error { return nil })
	if registry.Len() != 1 {
		t.Fatalf("expected one handler, got %d", registry.Len())
	}
}

// TestNewDecodesAndDispatchesCatalogPage verifies packet adapter behavior.
func TestNewDecodesAndDispatchesCatalogPage(t *testing.T) {
	packet, _ := codec.NewPacket(inpage.Header, inpage.Definition, codec.Int32(2), codec.Int32(-1), codec.String("NORMAL"))
	if err := New(pagecmd.Handler{}, zap.NewNop())(netconn.Context{}, packet); err == nil {
		t.Fatal("expected unresolved session error after dispatch")
	}
	if err := New(pagecmd.Handler{}, zap.NewNop())(netconn.Context{}, codec.Packet{Header: inpage.Header}); err == nil {
		t.Fatal("expected malformed packet error")
	}
}
