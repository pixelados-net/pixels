package room

import (
	permissionservice "github.com/niflaot/pixels/internal/permission/service"
	realmconn "github.com/niflaot/pixels/internal/realm/connection"
	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	playerservice "github.com/niflaot/pixels/internal/realm/player/service"
	respondcmd "github.com/niflaot/pixels/internal/realm/room/access/commands/doorbell/respond"
	entercmd "github.com/niflaot/pixels/internal/realm/room/access/commands/enter"
	entrytilecmd "github.com/niflaot/pixels/internal/realm/room/access/commands/entrytile"
	leavecmd "github.com/niflaot/pixels/internal/realm/room/access/commands/leave"
	modelcmd "github.com/niflaot/pixels/internal/realm/room/access/commands/model"
	roomentry "github.com/niflaot/pixels/internal/realm/room/access/entry"
	entryhandler "github.com/niflaot/pixels/internal/realm/room/access/handlers/entry"
	floorplancmd "github.com/niflaot/pixels/internal/realm/room/control/commands/floorplan"
	roomfloorplan "github.com/niflaot/pixels/internal/realm/room/control/floorplan"
	floorplanhandler "github.com/niflaot/pixels/internal/realm/room/control/handlers/floorplan"
	roommoderation "github.com/niflaot/pixels/internal/realm/room/control/moderation"
	roomrights "github.com/niflaot/pixels/internal/realm/room/control/rights"
	roomsettings "github.com/niflaot/pixels/internal/realm/room/control/settings"
	roomvotes "github.com/niflaot/pixels/internal/realm/room/control/votes"
	roomwordfilter "github.com/niflaot/pixels/internal/realm/room/control/wordfilter"
	tagscmd "github.com/niflaot/pixels/internal/realm/room/record/commands/tags"
	roomservice "github.com/niflaot/pixels/internal/realm/room/record/service"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	actionservice "github.com/niflaot/pixels/internal/realm/room/world/action"
	actioncmd "github.com/niflaot/pixels/internal/realm/room/world/commands/action"
	handitemcmd "github.com/niflaot/pixels/internal/realm/room/world/commands/handitem"
	lookcmd "github.com/niflaot/pixels/internal/realm/room/world/commands/look"
	walkcmd "github.com/niflaot/pixels/internal/realm/room/world/commands/walk"
	actionhandler "github.com/niflaot/pixels/internal/realm/room/world/handlers/action"
	handitemhandler "github.com/niflaot/pixels/internal/realm/room/world/handlers/handitem"
	movementhandler "github.com/niflaot/pixels/internal/realm/room/world/handlers/movement"
	"github.com/niflaot/pixels/internal/realm/room/world/layout"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	netconn "github.com/niflaot/pixels/networking/connection"
	"github.com/niflaot/pixels/pkg/bus"
	"github.com/niflaot/pixels/pkg/i18n"
	"github.com/niflaot/pixels/pkg/redis"
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
	// RoomLayouts resolves and persists room-owned layouts.
	RoomLayouts layout.RoomManager
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
	// Floorplan authorizes room floor plan changes.
	Floorplan *roomfloorplan.Authorizer
	// FloorplanConfig stores floor plan limits and throttling.
	FloorplanConfig roomfloorplan.Config
	// Redis stores distributed floor plan save cooldowns.
	Redis *redis.Client
	// WordFilters manages room-specific chat filter words.
	WordFilters roomwordfilter.Manager
	// Votes manages durable room upvotes.
	Votes roomvotes.Manager
	// Translations resolves end-user room control messages.
	Translations i18n.Translator
	// Actions changes live avatar expressions and posture.
	Actions *actionservice.Service
}

// RegisterConnectionHandlers registers room packet handlers.
func RegisterConnectionHandlers(handlers *realmconn.Handlers, deps HandlerDeps) {
	if handlers == nil || handlers.Inbound == nil {
		return
	}
	registerRightsHandlers(handlers.Inbound, deps)
	registerModerationHandlers(handlers.Inbound, deps)
	registerSettingsHandlers(handlers.Inbound, deps)
	registerVoteHandlers(handlers.Inbound, deps)
	registerFloorplanHandlers(handlers.Inbound, deps)

	enterCommand := newEnterCommand(deps)
	entryhandler.RegisterEnter(handlers.Inbound, entryhandler.NewEnter(enterCommand, deps.Log))
	entryhandler.RegisterDoorbell(handlers.Inbound, entryhandler.NewDoorbell(respondcmd.Handler{
		Players: deps.Players, Bindings: deps.Bindings, Runtime: deps.Runtime,
		Connections: deps.Connections, Entry: deps.Entry,
		Enter: enterCommand,
	}, deps.Log))
	entryhandler.RegisterModel(handlers.Inbound, entryhandler.NewModel(modelcmd.Handler{
		Players: deps.Players, Bindings: deps.Bindings, Rooms: deps.Rooms, Layouts: deps.Layouts,
	}, deps.Log))
	entryhandler.RegisterEntryTile(handlers.Inbound, entryhandler.NewEntryTile(entrytilecmd.Handler{
		Players: deps.Players, Bindings: deps.Bindings, Rooms: deps.Rooms, Layouts: deps.Layouts,
	}, deps.Log))
	entryhandler.RegisterTags(handlers.Inbound, entryhandler.NewTags(tagscmd.Handler{
		Players: deps.Players, Bindings: deps.Bindings, Rooms: deps.Rooms,
	}, deps.Log))
	movementhandler.RegisterLook(handlers.Inbound, movementhandler.NewLook(lookcmd.Handler{
		Players: deps.Players, Bindings: deps.Bindings, Runtime: deps.Runtime, Connections: deps.Connections,
	}, deps.Log))
	movementhandler.RegisterWalk(handlers.Inbound, movementhandler.NewWalk(walkcmd.Handler{
		Players: deps.Players, Bindings: deps.Bindings, Runtime: deps.Runtime, Connections: deps.Connections, Actions: deps.Actions,
	}, deps.Log))
	actionhandler.Register(handlers.Inbound, actionhandler.New(actioncmd.Handler{
		Players: deps.Players, Bindings: deps.Bindings, Runtime: deps.Runtime, Actions: deps.Actions,
	}, deps.Log))
	handitemhandler.Register(handlers.Inbound, handitemcmd.Handler{
		Players: deps.Players, Bindings: deps.Bindings, Runtime: deps.Runtime, Connections: deps.Connections,
	}, deps.Log)
	entryhandler.RegisterDesktop(handlers.Inbound, entryhandler.NewDesktop(leavecmd.Handler{
		Players: deps.Players, Bindings: deps.Bindings, Runtime: deps.Runtime,
		Connections: deps.Connections, Events: deps.Events,
	}, deps.Log))
}

// registerFloorplanHandlers registers floor plan editor packet adapters.
func registerFloorplanHandlers(registry *netconn.HandlerRegistry, deps HandlerDeps) {
	floorplanhandler.RegisterBlockedTiles(registry, floorplanhandler.NewBlockedTiles(floorplancmd.BlockedTilesHandler{
		Players: deps.Players, Bindings: deps.Bindings, Rooms: deps.Rooms,
		Runtime: deps.Runtime, Authorize: deps.Floorplan,
	}, deps.Log))
	floorplanhandler.RegisterSave(registry, floorplanhandler.NewSave(floorplancmd.SaveHandler{
		Players: deps.Players, Bindings: deps.Bindings, Rooms: deps.Rooms,
		Layouts: deps.RoomLayouts, Furniture: deps.Furniture, Runtime: deps.Runtime,
		Connections: deps.Connections, Authorize: deps.Floorplan, Cooldowns: deps.Redis,
		Config: deps.FloorplanConfig, Events: deps.Events, Translations: deps.Translations, Log: deps.Log,
	}, deps.Log))
}

// newEnterCommand composes room entry behavior and its controller projection.
func newEnterCommand(deps HandlerDeps) entercmd.Handler {
	return entercmd.Handler{
		Players: deps.Players, Bindings: deps.Bindings, Rooms: deps.Rooms,
		Layouts: deps.Layouts, Furniture: deps.Furniture, PlayerDirectory: deps.PlayerDirectory,
		Runtime: deps.Runtime, Connections: deps.Connections, Events: deps.Events,
		Entry: deps.Entry, Rights: deps.Rights, Moderation: deps.Moderation,
		Votes: deps.Votes,
		Control: entercmd.ControlPolicy{
			Permissions:    deps.Permissions,
			RightsAnyGrant: RightsAnyGrant, RightsAnyRevoke: RightsAnyRevoke,
			ModerationAnyKick: ModerationAnyKick, ModerationAnyMute: ModerationAnyMute,
			ModerationAnyBan: ModerationAnyBan,
		},
	}
}
