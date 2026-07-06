package navigator

import (
	realmconn "github.com/niflaot/pixels/internal/realm/connection"
	cancreatecmd "github.com/niflaot/pixels/internal/realm/navigator/commands/cancreate"
	createcmd "github.com/niflaot/pixels/internal/realm/navigator/commands/create"
	initcmd "github.com/niflaot/pixels/internal/realm/navigator/commands/init"
	searchcmd "github.com/niflaot/pixels/internal/realm/navigator/commands/search"
	cancreatehandler "github.com/niflaot/pixels/internal/realm/navigator/handlers/cancreate"
	createhandler "github.com/niflaot/pixels/internal/realm/navigator/handlers/create"
	inithandler "github.com/niflaot/pixels/internal/realm/navigator/handlers/init"
	searchhandler "github.com/niflaot/pixels/internal/realm/navigator/handlers/search"
	navservice "github.com/niflaot/pixels/internal/realm/navigator/service"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	roomlive "github.com/niflaot/pixels/internal/realm/room/live"
	roomservice "github.com/niflaot/pixels/internal/realm/room/service"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	"github.com/niflaot/pixels/pkg/bus"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// RegisterConnectionHandlers registers navigator packet handlers.
func RegisterConnectionHandlers(handlers *realmconn.Handlers, deps HandlerDeps) {
	if handlers == nil || handlers.Inbound == nil {
		return
	}

	inithandler.Register(handlers.Inbound, inithandler.New(initcmd.Handler{
		Players:   deps.Players,
		Bindings:  deps.Bindings,
		Navigator: deps.Navigator,
		Rooms:     deps.Rooms,
		Events:    deps.Events,
	}, deps.Log))
	searchhandler.Register(handlers.Inbound, searchhandler.New(searchcmd.Handler{
		Players:   deps.Players,
		Bindings:  deps.Bindings,
		Navigator: deps.Navigator,
		Rooms:     deps.Rooms,
		Runtime:   deps.Runtime,
		Events:    deps.Events,
	}, deps.Log))
	cancreatehandler.Register(handlers.Inbound, cancreatehandler.New(cancreatecmd.Handler{
		Players:  deps.Players,
		Bindings: deps.Bindings,
	}, deps.Log))
	createhandler.Register(handlers.Inbound, createhandler.New(createcmd.Handler{
		Players:  deps.Players,
		Bindings: deps.Bindings,
		Rooms:    deps.Rooms,
		Events:   deps.Events,
	}, deps.Log))
}

// HandlerDeps contains navigator handler dependencies.
type HandlerDeps struct {
	fx.In

	// Players stores live player state.
	Players *playerlive.Registry
	// Bindings stores player connection bindings.
	Bindings *binding.Registry
	// Navigator manages navigator persistence.
	Navigator navservice.Manager
	// Rooms manages room persistence.
	Rooms roomservice.Manager
	// Runtime stores active room runtime state.
	Runtime *roomlive.Registry
	// Events publishes realm events.
	Events *bus.Bus
	// Log records command dispatch.
	Log *zap.Logger
}
