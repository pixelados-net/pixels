package furniture

import (
	realmconn "github.com/niflaot/pixels/internal/realm/connection"
	inventorycmd "github.com/niflaot/pixels/internal/realm/furniture/commands/inventory"
	movecmd "github.com/niflaot/pixels/internal/realm/furniture/commands/move"
	pickupcmd "github.com/niflaot/pixels/internal/realm/furniture/commands/pickup"
	placecmd "github.com/niflaot/pixels/internal/realm/furniture/commands/place"
	inventoryhandler "github.com/niflaot/pixels/internal/realm/furniture/handlers/inventory"
	movehandler "github.com/niflaot/pixels/internal/realm/furniture/handlers/move"
	pickuphandler "github.com/niflaot/pixels/internal/realm/furniture/handlers/pickup"
	placehandler "github.com/niflaot/pixels/internal/realm/furniture/handlers/place"
	"github.com/niflaot/pixels/internal/realm/furniture/service"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	netconn "github.com/niflaot/pixels/networking/connection"
	"github.com/niflaot/pixels/pkg/bus"
	"github.com/niflaot/pixels/pkg/i18n"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// HandlerDeps contains furniture handler dependencies.
type HandlerDeps struct {
	fx.In

	// Players stores live player state.
	Players *playerlive.Registry
	// Bindings stores player connection bindings.
	Bindings *binding.Registry
	// Furniture manages placed and inventory furniture records.
	Furniture service.Manager
	// Runtime stores active rooms.
	Runtime *roomlive.Registry
	// Connections stores active network connections.
	Connections *netconn.Registry
	// Events publishes furniture lifecycle events.
	Events *bus.Bus
	// Translations resolves end-user messages.
	Translations i18n.Translator
	// Log records command dispatch.
	Log *zap.Logger
}

// RegisterConnectionHandlers registers furniture packet handlers.
func RegisterConnectionHandlers(handlers *realmconn.Handlers, deps HandlerDeps) {
	if handlers == nil || handlers.Inbound == nil {
		return
	}

	inventoryhandler.Register(handlers.Inbound, inventoryhandler.New(inventorycmd.Handler{
		Players: deps.Players, Bindings: deps.Bindings, Furniture: deps.Furniture,
	}, deps.Log))
	placehandler.Register(handlers.Inbound, placehandler.New(placecmd.Handler{
		Players: deps.Players, Bindings: deps.Bindings, Furniture: deps.Furniture,
		Runtime: deps.Runtime, Connections: deps.Connections, Events: deps.Events, Translations: deps.Translations, Log: deps.Log,
	}, deps.Log))
	movehandler.Register(handlers.Inbound, movehandler.New(movecmd.Handler{
		Players: deps.Players, Bindings: deps.Bindings, Furniture: deps.Furniture,
		Runtime: deps.Runtime, Connections: deps.Connections, Events: deps.Events, Translations: deps.Translations, Log: deps.Log,
	}, deps.Log))
	pickuphandler.Register(handlers.Inbound, pickuphandler.New(pickupcmd.Handler{
		Players: deps.Players, Bindings: deps.Bindings, Furniture: deps.Furniture,
		Runtime: deps.Runtime, Connections: deps.Connections, Events: deps.Events, Translations: deps.Translations, Log: deps.Log,
	}, deps.Log))
}
