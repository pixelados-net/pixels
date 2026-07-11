package navigator

import (
	realmconn "github.com/niflaot/pixels/internal/realm/connection"
	cancreatecmd "github.com/niflaot/pixels/internal/realm/navigator/commands/cancreate"
	countscmd "github.com/niflaot/pixels/internal/realm/navigator/commands/categorycounts"
	createcmd "github.com/niflaot/pixels/internal/realm/navigator/commands/create"
	eventcatscmd "github.com/niflaot/pixels/internal/realm/navigator/commands/eventcats"
	flatcatscmd "github.com/niflaot/pixels/internal/realm/navigator/commands/flatcats"
	forwardcmd "github.com/niflaot/pixels/internal/realm/navigator/commands/forward"
	infocmd "github.com/niflaot/pixels/internal/realm/navigator/commands/info"
	initcmd "github.com/niflaot/pixels/internal/realm/navigator/commands/init"
	searchcmd "github.com/niflaot/pixels/internal/realm/navigator/commands/search"
	cancreatehandler "github.com/niflaot/pixels/internal/realm/navigator/handlers/cancreate"
	countshandler "github.com/niflaot/pixels/internal/realm/navigator/handlers/categorycounts"
	createhandler "github.com/niflaot/pixels/internal/realm/navigator/handlers/create"
	eventcatshandler "github.com/niflaot/pixels/internal/realm/navigator/handlers/eventcats"
	flatcatshandler "github.com/niflaot/pixels/internal/realm/navigator/handlers/flatcats"
	forwardhandler "github.com/niflaot/pixels/internal/realm/navigator/handlers/forward"
	infohandler "github.com/niflaot/pixels/internal/realm/navigator/handlers/info"
	inithandler "github.com/niflaot/pixels/internal/realm/navigator/handlers/init"
	searchhandler "github.com/niflaot/pixels/internal/realm/navigator/handlers/search"
	navservice "github.com/niflaot/pixels/internal/realm/navigator/service"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	roommoderation "github.com/niflaot/pixels/internal/realm/room/control/moderation"
	roomrights "github.com/niflaot/pixels/internal/realm/room/control/rights"
	roomservice "github.com/niflaot/pixels/internal/realm/room/record/service"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	"github.com/niflaot/pixels/pkg/bus"
	"github.com/niflaot/pixels/pkg/i18n"
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
		Rights:    deps.Rights,
		Events:    deps.Events,
	}, deps.Log))
	cancreatehandler.Register(handlers.Inbound, cancreatehandler.New(cancreatecmd.Handler{
		Players:  deps.Players,
		Bindings: deps.Bindings,
		Rooms:    deps.Rooms,
	}, deps.Log))
	createhandler.Register(handlers.Inbound, createhandler.New(createcmd.Handler{
		Players:      deps.Players,
		Bindings:     deps.Bindings,
		Rooms:        deps.Rooms,
		Events:       deps.Events,
		Translations: deps.Translations,
		Log:          deps.Log,
	}, deps.Log))
	infohandler.Register(handlers.Inbound, infohandler.New(infocmd.Handler{
		Players:    deps.Players,
		Bindings:   deps.Bindings,
		Rooms:      deps.Rooms,
		Runtime:    deps.Runtime,
		Moderation: deps.Moderation,
	}, deps.Log))
	forwardhandler.Register(handlers.Inbound, forwardhandler.New(forwardcmd.Handler{
		Players:  deps.Players,
		Bindings: deps.Bindings,
		Rooms:    deps.Rooms,
	}, deps.Log))
	flatcatshandler.Register(handlers.Inbound, flatcatshandler.New(flatcatscmd.Handler{
		Players:    deps.Players,
		Bindings:   deps.Bindings,
		Categories: deps.Rooms,
	}, deps.Log))
	eventcatshandler.Register(handlers.Inbound, eventcatshandler.New(eventcatscmd.Handler{
		Players:  deps.Players,
		Bindings: deps.Bindings,
	}, deps.Log))
	countshandler.Register(handlers.Inbound, countshandler.New(countscmd.Handler{
		Players:  deps.Players,
		Bindings: deps.Bindings,
		Counts:   deps.Counts,
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
	// Rights resolves persistent room-scoped visibility.
	Rights roomrights.Manager
	// Moderation resolves viewer room moderation capability.
	Moderation roommoderation.Manager
	// Counts stores current navigator category counts.
	Counts *CategoryCountBroadcaster
	// Events publishes realm events.
	Events *bus.Bus
	// Log records command dispatch.
	Log *zap.Logger
	// Translations resolves localized navigator feedback.
	Translations i18n.Translator
}
