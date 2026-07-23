package navigator

import (
	permissionservice "github.com/niflaot/pixels/internal/permission/service"
	realmconn "github.com/niflaot/pixels/internal/realm/connection"
	socialgroup "github.com/niflaot/pixels/internal/realm/group"
	messengercore "github.com/niflaot/pixels/internal/realm/messenger/core"
	countscmd "github.com/niflaot/pixels/internal/realm/navigator/browse/category/counts"
	eventcatscmd "github.com/niflaot/pixels/internal/realm/navigator/browse/category/events"
	flatcatscmd "github.com/niflaot/pixels/internal/realm/navigator/browse/category/list"
	forwardcmd "github.com/niflaot/pixels/internal/realm/navigator/browse/room/forward"
	globalidcmd "github.com/niflaot/pixels/internal/realm/navigator/browse/room/globalid"
	infocmd "github.com/niflaot/pixels/internal/realm/navigator/browse/room/info"
	navruntime "github.com/niflaot/pixels/internal/realm/navigator/browse/runtime"
	searchcmd "github.com/niflaot/pixels/internal/realm/navigator/browse/search"
	navservice "github.com/niflaot/pixels/internal/realm/navigator/core"
	cancreatecmd "github.com/niflaot/pixels/internal/realm/navigator/create/check"
	createcmd "github.com/niflaot/pixels/internal/realm/navigator/create/room"
	favoritecmd "github.com/niflaot/pixels/internal/realm/navigator/favorite"
	navsession "github.com/niflaot/pixels/internal/realm/navigator/session"
	initcmd "github.com/niflaot/pixels/internal/realm/navigator/session/init"
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

	initcmd.RegisterPacketHandler(handlers.Inbound, initcmd.NewPacketHandler(initcmd.Handler{
		Players:   deps.Players,
		Bindings:  deps.Bindings,
		Navigator: deps.Navigator,
		Rooms:     deps.Rooms,
		Events:    deps.Events,
	}, deps.Log))
	searchcmd.RegisterPacketHandler(handlers.Inbound, searchcmd.NewPacketHandler(searchcmd.Handler{
		Players:      deps.Players,
		Bindings:     deps.Bindings,
		Navigator:    deps.Navigator,
		Rooms:        deps.Rooms,
		Runtime:      deps.Runtime,
		Events:       deps.Events,
		Groups:       deps.Groups,
		Friends:      deps.Messenger,
		GroupRooms:   deps.Groups,
		RightRooms:   deps.RightsService,
		Limit:        deps.Config.SearchLimit,
		HistoryLimit: deps.Config.HistoryLimit,
	}, deps.Log))
	searchcmd.RegisterLegacyAliases(handlers.Inbound, searchcmd.NewLegacyAliasHandler())
	cancreatecmd.RegisterPacketHandler(handlers.Inbound, cancreatecmd.NewPacketHandler(cancreatecmd.Handler{
		Players:  deps.Players,
		Bindings: deps.Bindings,
		Rooms:    deps.Rooms,
	}, deps.Log))
	createcmd.RegisterPacketHandler(handlers.Inbound, createcmd.NewPacketHandler(createcmd.Handler{
		Players:      deps.Players,
		Bindings:     deps.Bindings,
		Rooms:        deps.Rooms,
		Events:       deps.Events,
		Translations: deps.Translations,
		Log:          deps.Log,
	}, deps.Log))
	infocmd.RegisterPacketHandler(handlers.Inbound, infocmd.NewPacketHandler(infocmd.Handler{
		Players:    deps.Players,
		Bindings:   deps.Bindings,
		Rooms:      deps.Rooms,
		Runtime:    deps.Runtime,
		Moderation: deps.Moderation,
		Groups:     deps.Groups,
	}, deps.Log))
	forwardcmd.RegisterPacketHandler(handlers.Inbound, forwardcmd.NewPacketHandler(forwardcmd.Handler{
		Players:  deps.Players,
		Bindings: deps.Bindings,
		Rooms:    deps.Rooms,
	}, deps.Log))
	flatcatscmd.RegisterPacketHandler(handlers.Inbound, flatcatscmd.NewPacketHandler(flatcatscmd.Handler{
		Players:    deps.Players,
		Bindings:   deps.Bindings,
		Categories: deps.Rooms,
	}, deps.Log))
	eventcatscmd.RegisterPacketHandler(handlers.Inbound, eventcatscmd.NewPacketHandler(eventcatscmd.Handler{
		Players:  deps.Players,
		Bindings: deps.Bindings,
	}, deps.Log))
	countscmd.RegisterPacketHandler(handlers.Inbound, countscmd.NewPacketHandler(countscmd.Handler{
		Players:  deps.Players,
		Bindings: deps.Bindings,
		Counts:   deps.Counts,
	}, deps.Log))
	favoritecmd.RegisterHandlers(handlers.Inbound, favoritecmd.Handler{Navigator: deps.Navigator, Players: deps.Players, Bindings: deps.Bindings, Events: deps.Events, Rooms: deps.Rooms, Rights: deps.Rights, Permissions: deps.Permissions, Limit: deps.Config.FavoriteLimit, Translations: deps.Translations})
	navsession.RegisterSettings(handlers.Inbound, navsession.SettingsHandler{Navigator: deps.Navigator, Writer: deps.PreferenceWriter, Players: deps.Players, Bindings: deps.Bindings, PositionLimit: deps.Config.WindowPositionLimit, MinimumWidth: deps.Config.WindowMinimumWidth, MaximumWidth: deps.Config.WindowMaximumWidth, MinimumHeight: deps.Config.WindowMinimumHeight, MaximumHeight: deps.Config.WindowMaximumHeight})
	globalidcmd.Register(handlers.Inbound, globalidcmd.Handler{Players: deps.Players, Bindings: deps.Bindings, Rooms: deps.Rooms, Rights: deps.Rights})
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
	Counts *navruntime.CategoryCountBroadcaster
	// Events publishes realm events.
	Events *bus.Bus
	// Log records command dispatch.
	Log *zap.Logger
	// Translations resolves localized navigator feedback.
	Translations i18n.Translator
	// Groups resolves batched social-group room metadata.
	Groups *socialgroup.Service
	// Messenger reads cached directional friendship identifiers.
	Messenger *messengercore.Service
	// RightsService reads explicit rights room identifiers.
	RightsService *roomrights.Service
	// Permissions resolves Navigator-specific permission bypasses.
	Permissions permissionservice.Checker
	// Config contains bounded Navigator policy.
	Config Config
	// PreferenceWriter coalesces repeated resize settings.
	PreferenceWriter *navsession.PreferenceWriter
}
