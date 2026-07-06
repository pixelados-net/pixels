package room

import (
	realmconn "github.com/niflaot/pixels/internal/realm/connection"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	entercmd "github.com/niflaot/pixels/internal/realm/room/commands/enter"
	modelcmd "github.com/niflaot/pixels/internal/realm/room/commands/model"
	tagscmd "github.com/niflaot/pixels/internal/realm/room/commands/tags"
	enterhandler "github.com/niflaot/pixels/internal/realm/room/handlers/enter"
	modelhandler "github.com/niflaot/pixels/internal/realm/room/handlers/model"
	tagshandler "github.com/niflaot/pixels/internal/realm/room/handlers/tags"
	"github.com/niflaot/pixels/internal/realm/room/layout"
	roomlive "github.com/niflaot/pixels/internal/realm/room/live"
	roomservice "github.com/niflaot/pixels/internal/realm/room/service"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	"github.com/niflaot/pixels/pkg/bus"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// HandlerDeps contains room handler dependencies.
type HandlerDeps struct {
	fx.In

	// Players stores live player state.
	Players *playerlive.Registry
	// Bindings stores player connection bindings.
	Bindings *binding.Registry
	// Rooms manages room persistence.
	Rooms roomservice.Manager
	// Layouts manages room layouts.
	Layouts layout.Manager
	// Runtime stores active rooms.
	Runtime *roomlive.Registry
	// Events publishes realm events.
	Events *bus.Bus
	// Log records command dispatch.
	Log *zap.Logger
}

// RegisterConnectionHandlers registers room packet handlers.
func RegisterConnectionHandlers(handlers *realmconn.Handlers, deps HandlerDeps) {
	if handlers == nil || handlers.Inbound == nil {
		return
	}

	enterhandler.Register(handlers.Inbound, enterhandler.New(entercmd.Handler{
		Players: deps.Players, Bindings: deps.Bindings, Rooms: deps.Rooms,
		Layouts: deps.Layouts, Runtime: deps.Runtime, Events: deps.Events,
	}, deps.Log))
	modelhandler.Register(handlers.Inbound, modelhandler.New(modelcmd.Handler{
		Players: deps.Players, Bindings: deps.Bindings, Rooms: deps.Rooms, Layouts: deps.Layouts,
	}, deps.Log))
	tagshandler.Register(handlers.Inbound, tagshandler.New(tagscmd.Handler{
		Players: deps.Players, Bindings: deps.Bindings, Rooms: deps.Rooms,
	}, deps.Log))
}
