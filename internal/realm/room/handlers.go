package room

import (
	permissionservice "github.com/niflaot/pixels/internal/permission/service"
	realmconn "github.com/niflaot/pixels/internal/realm/connection"
	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	playerservice "github.com/niflaot/pixels/internal/realm/player/service"
	respondcmd "github.com/niflaot/pixels/internal/realm/room/commands/doorbell/respond"
	entercmd "github.com/niflaot/pixels/internal/realm/room/commands/enter"
	entrytilecmd "github.com/niflaot/pixels/internal/realm/room/commands/entrytile"
	leavecmd "github.com/niflaot/pixels/internal/realm/room/commands/leave"
	lookcmd "github.com/niflaot/pixels/internal/realm/room/commands/look"
	modelcmd "github.com/niflaot/pixels/internal/realm/room/commands/model"
	tagscmd "github.com/niflaot/pixels/internal/realm/room/commands/tags"
	walkcmd "github.com/niflaot/pixels/internal/realm/room/commands/walk"
	roomentry "github.com/niflaot/pixels/internal/realm/room/entry"
	desktophandler "github.com/niflaot/pixels/internal/realm/room/handlers/desktop"
	respondhandler "github.com/niflaot/pixels/internal/realm/room/handlers/doorbell/respond"
	enterhandler "github.com/niflaot/pixels/internal/realm/room/handlers/enter"
	entrytilehandler "github.com/niflaot/pixels/internal/realm/room/handlers/entrytile"
	lookhandler "github.com/niflaot/pixels/internal/realm/room/handlers/look"
	modelhandler "github.com/niflaot/pixels/internal/realm/room/handlers/model"
	tagshandler "github.com/niflaot/pixels/internal/realm/room/handlers/tags"
	walkhandler "github.com/niflaot/pixels/internal/realm/room/handlers/walk"
	"github.com/niflaot/pixels/internal/realm/room/layout"
	roomlive "github.com/niflaot/pixels/internal/realm/room/live"
	roommoderation "github.com/niflaot/pixels/internal/realm/room/moderation"
	roomrights "github.com/niflaot/pixels/internal/realm/room/rights"
	roomservice "github.com/niflaot/pixels/internal/realm/room/service"
	roomsettings "github.com/niflaot/pixels/internal/realm/room/settings"
	roomwordfilter "github.com/niflaot/pixels/internal/realm/room/wordfilter"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	netconn "github.com/niflaot/pixels/networking/connection"
	"github.com/niflaot/pixels/pkg/bus"
	"github.com/niflaot/pixels/pkg/i18n"
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
	// ConfigRooms persists focused room settings updates.
	ConfigRooms roomservice.ConfigManager
	// Layouts manages room layouts.
	Layouts layout.Manager
	// Furniture manages placed and inventory furniture records.
	Furniture furnitureservice.Manager
	// PlayerDirectory resolves durable player identities for furniture owners not currently online.
	PlayerDirectory playerservice.Finder
	// Runtime stores active rooms.
	Runtime *roomlive.Registry
	// Connections stores active network connections.
	Connections *netconn.Registry
	// Events publishes realm events.
	Events *bus.Bus
	// Log records command dispatch.
	Log *zap.Logger
	// Entry decides closed-room access.
	Entry *roomentry.Service
	// Rights manages persistent room build rights.
	Rights *roomrights.Service
	// Permissions resolves global room control capabilities.
	Permissions permissionservice.Checker
	// Moderation manages room sanctions and kicks.
	Moderation *roommoderation.Service
	// Settings authorizes room configuration changes.
	Settings *roomsettings.Authorizer
	// WordFilters manages room-specific chat filter words.
	WordFilters roomwordfilter.Manager
	// Translations resolves end-user room control messages.
	Translations i18n.Translator
}

// RegisterConnectionHandlers registers room packet handlers.
func RegisterConnectionHandlers(handlers *realmconn.Handlers, deps HandlerDeps) {
	if handlers == nil || handlers.Inbound == nil {
		return
	}
	registerRightsHandlers(handlers.Inbound, deps)
	registerModerationHandlers(handlers.Inbound, deps)
	registerSettingsHandlers(handlers.Inbound, deps)

	enterCommand := newEnterCommand(deps)
	enterhandler.Register(handlers.Inbound, enterhandler.New(enterCommand, deps.Log))
	respondhandler.Register(handlers.Inbound, respondhandler.New(respondcmd.Handler{
		Players: deps.Players, Bindings: deps.Bindings, Runtime: deps.Runtime,
		Connections: deps.Connections, Entry: deps.Entry,
		Enter: enterCommand,
	}, deps.Log))
	modelhandler.Register(handlers.Inbound, modelhandler.New(modelcmd.Handler{
		Players: deps.Players, Bindings: deps.Bindings, Rooms: deps.Rooms, Layouts: deps.Layouts,
	}, deps.Log))
	entrytilehandler.Register(handlers.Inbound, entrytilehandler.New(entrytilecmd.Handler{
		Players: deps.Players, Bindings: deps.Bindings, Rooms: deps.Rooms, Layouts: deps.Layouts,
	}, deps.Log))
	tagshandler.Register(handlers.Inbound, tagshandler.New(tagscmd.Handler{
		Players: deps.Players, Bindings: deps.Bindings, Rooms: deps.Rooms,
	}, deps.Log))
	lookhandler.Register(handlers.Inbound, lookhandler.New(lookcmd.Handler{
		Players: deps.Players, Bindings: deps.Bindings, Runtime: deps.Runtime, Connections: deps.Connections,
	}, deps.Log))
	walkhandler.Register(handlers.Inbound, walkhandler.New(walkcmd.Handler{
		Players: deps.Players, Bindings: deps.Bindings, Runtime: deps.Runtime, Connections: deps.Connections,
	}, deps.Log))
	desktophandler.Register(handlers.Inbound, desktophandler.New(leavecmd.Handler{
		Players: deps.Players, Bindings: deps.Bindings, Runtime: deps.Runtime,
		Connections: deps.Connections, Events: deps.Events,
	}, deps.Log))
}

// newEnterCommand composes room entry behavior and its controller projection.
func newEnterCommand(deps HandlerDeps) entercmd.Handler {
	return entercmd.Handler{
		Players: deps.Players, Bindings: deps.Bindings, Rooms: deps.Rooms,
		Layouts: deps.Layouts, Furniture: deps.Furniture, PlayerDirectory: deps.PlayerDirectory,
		Runtime: deps.Runtime, Connections: deps.Connections, Events: deps.Events,
		Entry: deps.Entry, Rights: deps.Rights,
		Control: entercmd.ControlPolicy{
			Permissions:    deps.Permissions,
			RightsAnyGrant: RightsAnyGrant, RightsAnyRevoke: RightsAnyRevoke,
			ModerationAnyKick: ModerationAnyKick, ModerationAnyMute: ModerationAnyMute,
			ModerationAnyBan: ModerationAnyBan,
		},
	}
}
