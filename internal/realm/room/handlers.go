package room

import (
	realmconn "github.com/niflaot/pixels/internal/realm/connection"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	entercmd "github.com/niflaot/pixels/internal/realm/room/commands/enter"
	leavecmd "github.com/niflaot/pixels/internal/realm/room/commands/leave"
	modelcmd "github.com/niflaot/pixels/internal/realm/room/commands/model"
	tagscmd "github.com/niflaot/pixels/internal/realm/room/commands/tags"
	walkcmd "github.com/niflaot/pixels/internal/realm/room/commands/walk"
	desktophandler "github.com/niflaot/pixels/internal/realm/room/handlers/desktop"
	enterhandler "github.com/niflaot/pixels/internal/realm/room/handlers/enter"
	modelhandler "github.com/niflaot/pixels/internal/realm/room/handlers/model"
	tagshandler "github.com/niflaot/pixels/internal/realm/room/handlers/tags"
	walkhandler "github.com/niflaot/pixels/internal/realm/room/handlers/walk"
	"github.com/niflaot/pixels/internal/realm/room/layout"
	roomlive "github.com/niflaot/pixels/internal/realm/room/live"
	roomservice "github.com/niflaot/pixels/internal/realm/room/service"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	netconn "github.com/niflaot/pixels/networking/connection"
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
	// Connections stores active network connections.
	Connections *netconn.Registry
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
		Layouts: deps.Layouts, Runtime: deps.Runtime, Connections: deps.Connections, Events: deps.Events,
	}, deps.Log))
	modelhandler.Register(handlers.Inbound, modelhandler.New(modelcmd.Handler{
		Players: deps.Players, Bindings: deps.Bindings, Rooms: deps.Rooms, Layouts: deps.Layouts,
	}, deps.Log))
	tagshandler.Register(handlers.Inbound, tagshandler.New(tagscmd.Handler{
		Players: deps.Players, Bindings: deps.Bindings, Rooms: deps.Rooms,
	}, deps.Log))
	walkhandler.Register(handlers.Inbound, walkhandler.New(walkcmd.Handler{
		Players: deps.Players, Bindings: deps.Bindings, Runtime: deps.Runtime,
	}, deps.Log))
	desktophandler.Register(handlers.Inbound, desktophandler.New(leavecmd.Handler{
		Players: deps.Players, Bindings: deps.Bindings, Runtime: deps.Runtime,
		Connections: deps.Connections, Events: deps.Events,
	}, deps.Log))
}
