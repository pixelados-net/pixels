// Package moderation wires call-for-help, staff tools, guides, and guardians.
package moderation

import (
	permissionservice "github.com/niflaot/pixels/internal/permission/service"
	camerareport "github.com/niflaot/pixels/internal/realm/camera/report"
	"github.com/niflaot/pixels/internal/realm/chat/history"
	realmconnection "github.com/niflaot/pixels/internal/realm/connection"
	"github.com/niflaot/pixels/internal/realm/moderation/cfh"
	moderationconfig "github.com/niflaot/pixels/internal/realm/moderation/config"
	moderationcore "github.com/niflaot/pixels/internal/realm/moderation/core"
	moderationdb "github.com/niflaot/pixels/internal/realm/moderation/database"
	"github.com/niflaot/pixels/internal/realm/moderation/guardian"
	guardianhandlers "github.com/niflaot/pixels/internal/realm/moderation/guardian/handlers"
	"github.com/niflaot/pixels/internal/realm/moderation/guide"
	guidehandlers "github.com/niflaot/pixels/internal/realm/moderation/guide/handlers"
	moderationrecord "github.com/niflaot/pixels/internal/realm/moderation/record"
	moderationruntime "github.com/niflaot/pixels/internal/realm/moderation/runtime"
	"github.com/niflaot/pixels/internal/realm/moderation/staff"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	playerservice "github.com/niflaot/pixels/internal/realm/player/service"
	roomservice "github.com/niflaot/pixels/internal/realm/room/record/service"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	sanctioncore "github.com/niflaot/pixels/internal/realm/sanction/core"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	netconn "github.com/niflaot/pixels/networking/connection"
	"github.com/niflaot/pixels/pkg/bus"
	"github.com/niflaot/pixels/pkg/i18n"
	"github.com/niflaot/pixels/pkg/redis"
	"go.uber.org/fx"
)

// Module provides global moderation behavior.
var Module = fx.Module(
	"realm-moderation",
	fx.Provide(
		moderationconfig.Load,
		moderationdb.New,
		NewStore,
		moderationcore.New,
		guide.New,
		NewGuardianManager,
		NewRuntime,
	),
	fx.Invoke(RegisterHandlers, RegisterLifecycle, RegisterAmbassadorIntake),
)

// NewStore exposes PostgreSQL persistence through the moderation boundary.
func NewStore(repository *moderationdb.Repository) moderationrecord.Store { return repository }

// NewGuardianManager creates peer review behavior over the live guide pool.
func NewGuardianManager(config moderationconfig.Config, guides *guide.Manager, redisClient *redis.Client, store moderationrecord.Store) *guardian.Manager {
	return guardian.NewPersistent(config, guides, redisClient, store)
}

// RuntimeDeps contains live moderation dependencies.
type RuntimeDeps struct {
	fx.In

	// Moderation coordinates reports and the issue queue.
	Moderation *moderationcore.Service
	// Sanctions applies hotel-wide punishments.
	Sanctions *sanctioncore.Service
	// Guides owns helper duty and sessions.
	Guides *guide.Manager
	// Guardians owns peer review tickets.
	Guardians *guardian.Manager
	// Players stores live player snapshots.
	Players *playerlive.Registry
	// PlayerRecords reads durable player profiles.
	PlayerRecords playerservice.Finder
	// Rooms reads and changes room settings.
	Rooms roomservice.ConfigManager
	// Runtime locates active rooms.
	Runtime *roomlive.Registry
	// History reads bounded room chat history.
	History *history.Service
	// Bindings resolves authenticated sources.
	Bindings *binding.Registry
	// Connections sends protocol responses.
	Connections *netconn.Registry
	// Permissions resolves moderation capabilities.
	Permissions permissionservice.Checker
	// Translations localizes user-visible text.
	Translations i18n.Translator
	// Events publishes completed guide lifecycle facts.
	Events bus.Publisher
}

// NewRuntime creates shared live moderation context.
func NewRuntime(deps RuntimeDeps) *moderationruntime.Context {
	return &moderationruntime.Context{Moderation: deps.Moderation, Sanctions: deps.Sanctions, Guides: deps.Guides, Guardians: deps.Guardians, Players: deps.Players, PlayerRecords: deps.PlayerRecords, Rooms: deps.Rooms, RoomsLive: deps.Runtime, History: deps.History, Bindings: deps.Bindings, Connections: deps.Connections, Permissions: deps.Permissions, Translations: deps.Translations, Events: deps.Events}
}

// RegisterHandlers installs all moderation protocol adapters.
func RegisterHandlers(connectionHandlers *realmconnection.Handlers, runtime *moderationruntime.Context, photos *camerareport.Service) {
	cfh.Register(connectionHandlers.Inbound, runtime, photos)
	staff.Register(connectionHandlers.Inbound, runtime)
	guidehandlers.Register(connectionHandlers.Inbound, runtime)
	guardianhandlers.Register(connectionHandlers.Inbound, runtime)
}
