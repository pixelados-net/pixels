package request

import (
	"testing"

	requestcmd "github.com/niflaot/pixels/internal/realm/inventory/currency/commands/request"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inrequest "github.com/niflaot/pixels/networking/inbound/user/currency/request"
)

// TestRegisterAddsCurrencyRequestHandler verifies packet routing.
func TestRegisterAddsCurrencyRequestHandler(t *testing.T) {
	registry := netconn.NewHandlerRegistry()
	Register(registry, New(&requestcmd.Handler{
		Players: playerlive.NewRegistry(), Bindings: binding.NewRegistry(),
	}, nil))

	err := registry.Handle(
		netconn.Context{State: netconn.StateConnected, Authenticated: true},
		codec.Packet{Header: inrequest.Header},
	)
	if err == nil {
		t.Fatal("expected missing command dependencies")
	}
}
