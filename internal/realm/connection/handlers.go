// Package connection contains connection-realm handlers and commands.
package connection

import (
	"context"

	"github.com/niflaot/pixels/internal/auth/sso"
	"github.com/niflaot/pixels/internal/realm/connection/handlers/handshake"
	"github.com/niflaot/pixels/internal/realm/connection/handlers/heartbeat"
	"github.com/niflaot/pixels/internal/realm/connection/handlers/latency"
	"github.com/niflaot/pixels/internal/realm/connection/handlers/security"
	currencyrequest "github.com/niflaot/pixels/internal/realm/inventory/currency/commands/request"
	playerdisconnected "github.com/niflaot/pixels/internal/realm/player/events/disconnected"
	"github.com/niflaot/pixels/internal/realm/player/live"
	playerservice "github.com/niflaot/pixels/internal/realm/player/service"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	sessionunbound "github.com/niflaot/pixels/internal/realm/session/events/unbound"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	"github.com/niflaot/pixels/pkg/bus"
)

// Handlers contains connection-realm handler registries.
type Handlers struct {
	// Inbound routes client packets.
	Inbound *netconn.HandlerRegistry
	// Outbound routes server packets.
	Outbound *netconn.HandlerRegistry
	// players stores live player runtime state.
	players *live.Registry
	// bindings stores player connection bindings.
	bindings *binding.Registry
	// events publishes realm lifecycle events.
	events bus.Publisher
}

// NewHandlers creates connection-realm handler registries.
func NewHandlers(sso *sso.Service, finder playerservice.Finder, players *live.Registry, bindings *binding.Registry, events *bus.Bus, currencies *currencyrequest.Handler) *Handlers {
	inbound := netconn.NewHandlerRegistry()
	outbound := netconn.NewHandlerRegistry()
	handlers := &Handlers{Inbound: inbound, Outbound: outbound, players: players, bindings: bindings, events: events}

	registerInbound(inbound, security.NewAuthenticator(sso, finder, players, bindings, events, currencies))
	outbound.SetFallback(noopHandler, netconn.AllowAnyActiveState(), netconn.AllowUnauthenticated())

	return handlers
}

// registerInbound registers connection-realm inbound handlers.
func registerInbound(registry *netconn.HandlerRegistry, authenticator *security.Authenticator) {
	handshake.Register(registry)
	security.Register(registry, authenticator)
	heartbeat.Register(registry)
	latency.Register(registry)
}

// Disconnected releases runtime player state for a disposed connection.
func (handlers *Handlers) Disconnected(ctx context.Context, kind netconn.Kind, id netconn.ID) {
	if handlers == nil || handlers.bindings == nil {
		return
	}

	sessionBinding, found := handlers.bindings.RemoveByConnection(binding.ConnectionKey{Kind: kind, ID: id})
	if !found {
		return
	}

	if handlers.players != nil {
		handlers.players.Remove(sessionBinding.PlayerID)
	}

	handlers.publish(ctx, sessionunbound.Name, sessionunbound.Payload{Binding: sessionBinding})
	handlers.publish(ctx, playerdisconnected.Name, playerdisconnected.Payload{
		PlayerID:       sessionBinding.PlayerID,
		ConnectionID:   sessionBinding.ConnectionID,
		ConnectionKind: sessionBinding.ConnectionKind,
	})
}

// publish emits an event when an event bus is configured.
func (handlers *Handlers) publish(ctx context.Context, name bus.Name, payload any) {
	if handlers.events == nil {
		return
	}

	_ = handlers.events.Publish(ctx, bus.Event{Name: name, Payload: payload})
}

// noopHandler accepts outbound packets without side effects.
func noopHandler(netconn.Context, codec.Packet) error {
	return nil
}
