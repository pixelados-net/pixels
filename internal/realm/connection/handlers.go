// Package connection contains connection-realm handlers and commands.
package connection

import (
	"context"

	"github.com/niflaot/pixels/internal/auth/sso"
	permissionbroadcast "github.com/niflaot/pixels/internal/permission/broadcast"
	protocolcompat "github.com/niflaot/pixels/internal/realm/connection/compatibility"
	"github.com/niflaot/pixels/internal/realm/connection/handlers/handshake"
	"github.com/niflaot/pixels/internal/realm/connection/handlers/heartbeat"
	"github.com/niflaot/pixels/internal/realm/connection/handlers/latency"
	"github.com/niflaot/pixels/internal/realm/connection/handlers/security"
	currencyrequest "github.com/niflaot/pixels/internal/realm/inventory/currency/commands/request"
	playerachievement "github.com/niflaot/pixels/internal/realm/player/achievement"
	playercompatibility "github.com/niflaot/pixels/internal/realm/player/compatibility"
	playereffect "github.com/niflaot/pixels/internal/realm/player/effect"
	playerdisconnected "github.com/niflaot/pixels/internal/realm/player/events/disconnected"
	playeridentity "github.com/niflaot/pixels/internal/realm/player/identity"
	"github.com/niflaot/pixels/internal/realm/player/live"
	playerprofile "github.com/niflaot/pixels/internal/realm/player/profile"
	playerservice "github.com/niflaot/pixels/internal/realm/player/service"
	playersettings "github.com/niflaot/pixels/internal/realm/player/settings"
	playerwardrobe "github.com/niflaot/pixels/internal/realm/player/wardrobe"
	roomrights "github.com/niflaot/pixels/internal/realm/room/control/rights"
	roomservice "github.com/niflaot/pixels/internal/realm/room/record/service"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	sessionunbound "github.com/niflaot/pixels/internal/realm/session/events/unbound"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	"github.com/niflaot/pixels/pkg/bus"
	"github.com/niflaot/pixels/pkg/i18n"
	"go.uber.org/zap"
)

// RegisterProtocolCompatibilityHandlers installs empty snapshots for optional client systems.
func RegisterProtocolCompatibilityHandlers(handlers *Handlers) {
	if handlers == nil {
		return
	}
	protocolcompat.Register(handlers.Inbound)
}

// RegisterEffectHandlers installs Nitro effect selection and activation adapters.
func RegisterEffectHandlers(handlers *Handlers, effects *playereffect.Service, bindings *binding.Registry, log *zap.Logger) {
	if handlers == nil {
		return
	}
	playereffect.RegisterHandlers(handlers.Inbound, playereffect.Handler{Effects: effects, Bindings: bindings, Log: log})
}

// RegisterAchievementHandlers installs Nitro badge inventory and projection adapters.
func RegisterAchievementHandlers(handlers *Handlers, achievements *playerachievement.Service, bindings *binding.Registry) {
	if handlers == nil {
		return
	}
	playerachievement.RegisterHandlers(handlers.Inbound, playerachievement.Handler{Achievements: achievements, Bindings: bindings})
}

// RegisterPlayerCompatibilityHandlers installs retired user protocol no-ops.
func RegisterPlayerCompatibilityHandlers(handlers *Handlers) {
	if handlers == nil {
		return
	}
	playercompatibility.Register(handlers.Inbound)
}

// RegisterSecurityTranslations installs localized authentication feedback.
func RegisterSecurityTranslations(handlers *Handlers, translations i18n.Translator) {
	if handlers == nil || handlers.authenticator == nil {
		return
	}
	handlers.authenticator.SetTranslations(translations)
}

// RegisterPlayerSettingsHandlers installs durable settings adapters and login hydration.
func RegisterPlayerSettingsHandlers(handlers *Handlers, settings *playersettings.Service, writer *playersettings.Writer, bindings *binding.Registry, players *live.Registry, rooms roomservice.Manager, rights *roomrights.Service, translations i18n.Translator) {
	if handlers == nil {
		return
	}
	playersettings.RegisterHandlers(handlers.Inbound, playersettings.Handler{Service: settings, Writer: writer, Bindings: bindings, Players: players, Rooms: rooms, Rights: rights, Translations: translations})
	if handlers.authenticator != nil {
		handlers.authenticator.SetSettingsLoader(settings)
	}
}

// RegisterPlayerProfileHandlers installs public profile adapters.
func RegisterPlayerProfileHandlers(handlers *Handlers, profiles *playerprofile.Service, bindings *binding.Registry, players *live.Registry, rooms *roomlive.Registry, connections *netconn.Registry, translations i18n.Translator, log *zap.Logger) {
	if handlers == nil {
		return
	}
	playerprofile.RegisterHandlers(handlers.Inbound, playerprofile.Handler{
		Service: profiles, Bindings: bindings, Players: players, Rooms: rooms, Connections: connections, Events: handlers.events, Translations: translations, Log: log,
	})
	if handlers.authenticator != nil {
		handlers.authenticator.SetRespectLoader(profiles)
	}
}

// RegisterPlayerWardrobeHandlers installs persistent wardrobe adapters.
func RegisterPlayerWardrobeHandlers(handlers *Handlers, wardrobe *playerwardrobe.Service, bindings *binding.Registry, translations i18n.Translator) {
	if handlers == nil {
		return
	}
	playerwardrobe.RegisterHandlers(handlers.Inbound, playerwardrobe.Handler{Service: wardrobe, Bindings: bindings, Translations: translations})
}

// RegisterPlayerIdentityHandlers installs username reservation and rename adapters.
func RegisterPlayerIdentityHandlers(handlers *Handlers, identities *playeridentity.Service, bindings *binding.Registry, players *live.Registry, rooms *roomlive.Registry, connections *netconn.Registry) {
	if handlers == nil {
		return
	}
	playeridentity.RegisterHandlers(handlers.Inbound, playeridentity.Handler{
		Service: identities, Bindings: bindings, Players: players, Rooms: rooms, Connections: connections, Events: handlers.events,
	})
}

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
	// authenticator resolves login tickets and global admission gates.
	authenticator *security.Authenticator
}

// NewHandlers creates connection-realm handler registries.
func NewHandlers(sso *sso.Service, finder playerservice.Finder, players *live.Registry, bindings *binding.Registry, events *bus.Bus, currencies *currencyrequest.Handler) *Handlers {
	return newHandlers(sso, finder, players, bindings, events, currencies, nil)
}

// NewHandlersWithPermissions creates handlers with permission bootstrap projection.
func NewHandlersWithPermissions(sso *sso.Service, finder playerservice.Finder, players *live.Registry, bindings *binding.Registry, events *bus.Bus, currencies *currencyrequest.Handler, permissions *permissionbroadcast.Projector) *Handlers {
	return newHandlers(sso, finder, players, bindings, events, currencies, permissions)
}

// newHandlers creates connection-realm handlers with optional permission projection.
func newHandlers(sso *sso.Service, finder playerservice.Finder, players *live.Registry, bindings *binding.Registry, events *bus.Bus, currencies *currencyrequest.Handler, permissions *permissionbroadcast.Projector) *Handlers {
	inbound := netconn.NewHandlerRegistry()
	outbound := netconn.NewHandlerRegistry()
	authenticator := security.NewAuthenticator(sso, finder, players, bindings, events, currencies, permissions)
	handlers := &Handlers{Inbound: inbound, Outbound: outbound, players: players, bindings: bindings, events: events, authenticator: authenticator}

	registerInbound(inbound, authenticator)
	outbound.SetFallback(noopHandler, netconn.AllowAnyActiveState(), netconn.AllowUnauthenticated())

	return handlers
}

// SetSanctionGate replaces the global login sanction validator.
func (handlers *Handlers) SetSanctionGate(gate security.SanctionGate) {
	if handlers == nil || handlers.authenticator == nil {
		return
	}
	handlers.authenticator.SetSanctionGate(gate)
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
