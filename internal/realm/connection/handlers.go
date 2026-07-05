// Package connection contains connection-realm handlers and commands.
package connection

import (
	"github.com/niflaot/pixels/internal/auth/sso"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inlatency "github.com/niflaot/pixels/networking/inbound/client/latency"
	inpong "github.com/niflaot/pixels/networking/inbound/client/pong"
	indiffiecomplete "github.com/niflaot/pixels/networking/inbound/handshake/diffie/complete"
	indiffieinit "github.com/niflaot/pixels/networking/inbound/handshake/diffie/init"
	inpolicy "github.com/niflaot/pixels/networking/inbound/handshake/policy"
	inrelease "github.com/niflaot/pixels/networking/inbound/handshake/release"
	invariables "github.com/niflaot/pixels/networking/inbound/handshake/variables"
	inmachine "github.com/niflaot/pixels/networking/inbound/security/machine"
	inticket "github.com/niflaot/pixels/networking/inbound/security/ticket"
)

// Handlers contains connection-realm handler registries.
type Handlers struct {
	// Inbound routes client packets.
	Inbound *netconn.HandlerRegistry
	// Outbound routes server packets.
	Outbound *netconn.HandlerRegistry
}

// NewHandlers creates connection-realm handler registries.
func NewHandlers(sso *sso.Service) *Handlers {
	inbound := netconn.NewHandlerRegistry()
	outbound := netconn.NewHandlerRegistry()
	handlers := &Handlers{Inbound: inbound, Outbound: outbound}

	registerInbound(inbound, sso)
	outbound.SetFallback(noopHandler, netconn.AllowAnyActiveState(), netconn.AllowUnauthenticated())

	return handlers
}

// registerInbound registers connection-realm inbound handlers.
func registerInbound(registry *netconn.HandlerRegistry, service *sso.Service) {
	early := []netconn.HandlerOption{netconn.AllowStates(netconn.StateCreated, netconn.StateHandshaking), netconn.AllowUnauthenticated()}
	_ = registry.Register(inrelease.Header, releaseHandler, early...)
	_ = registry.Register(invariables.Header, variablesHandler, early...)
	_ = registry.Register(inpolicy.Header, policyHandler, early...)
	_ = registry.Register(indiffieinit.Header, diffieInitHandler, early...)
	_ = registry.Register(indiffiecomplete.Header, diffieCompleteHandler, netconn.AllowStates(netconn.StateSecuring), netconn.AllowUnauthenticated())
	_ = registry.Register(inmachine.Header, machineHandler, netconn.AllowStates(netconn.StateHandshaking, netconn.StateSecuring), netconn.AllowUnauthenticated())
	_ = registry.Register(inticket.Header, ticketHandler(service), netconn.AllowStates(netconn.StateHandshaking), netconn.AllowUnauthenticated())
	_ = registry.Register(inpong.Header, pongHandler)
	_ = registry.Register(inlatency.Header, latencyHandler)
}

// noopHandler accepts outbound packets without side effects.
func noopHandler(netconn.Context, codec.Packet) error {
	return nil
}
