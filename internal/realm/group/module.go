package group

import (
	"context"

	realmconn "github.com/niflaot/pixels/internal/realm/connection"
	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	"github.com/niflaot/pixels/internal/realm/group/badge"
	groupconfig "github.com/niflaot/pixels/internal/realm/group/config"
	groupdb "github.com/niflaot/pixels/internal/realm/group/database"
	"github.com/niflaot/pixels/internal/realm/group/forum"
	"github.com/niflaot/pixels/internal/realm/group/identity"
	"github.com/niflaot/pixels/internal/realm/group/membership"
	groupobservability "github.com/niflaot/pixels/internal/realm/group/observability"
	grouprecord "github.com/niflaot/pixels/internal/realm/group/record"
	grouproom "github.com/niflaot/pixels/internal/realm/group/room"
	groupruntime "github.com/niflaot/pixels/internal/realm/group/runtime"
	moderationcore "github.com/niflaot/pixels/internal/realm/moderation/core"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/pkg/i18n"
	"go.uber.org/fx"
)

// Module provides social-group persistence, behavior, snapshots, and packet routing.
var Module = fx.Module(
	"realm-group",
	fx.Provide(
		groupconfig.LoadConfig,
		groupdb.New,
		groupobservability.New,
		NewStore,
		NewRecordStore,
		NewBadgeSource,
		NewWiredService,
		NewFurnitureGroupPolicy,
		groupruntime.NewCache,
		groupruntime.NewDelivery,
		groupruntime.NewProjector,
		badge.New,
		badge.NewCompiler,
		identity.New,
		membership.New,
		forum.New,
		forum.NewCursors,
	),
	fx.Invoke(RegisterObservability, RegisterLifecycle, RegisterSnapshots, RegisterRoomPolicy, RegisterConnectionHandlers),
)

// RegisterObservability attaches bounded telemetry to warmed group components.
func RegisterObservability(config groupconfig.Config, metrics *groupobservability.Metrics, compiler *badge.Compiler, cache *groupruntime.Cache, projector *groupruntime.Projector) {
	compiler.SetMetrics(metrics)
	cache.SetMetrics(metrics)
	cache.SetTTL(config.CacheTTL)
	projector.SetMetrics(metrics)
}

// NewFurnitureGroupPolicy exposes warmed linked furniture identity to placement projections.
func NewFurnitureGroupPolicy(service *Service) furnituremodel.GroupPolicy { return service }

// NewStore exposes PostgreSQL group reads through the WIRED boundary.
func NewStore(repository *groupdb.Repository) Store { return repository }

// NewRecordStore exposes complete PostgreSQL behavior through the domain boundary.
func NewRecordStore(repository *groupdb.Repository) grouprecord.Store { return repository }

// NewBadgeSource exposes immutable badge reference reads.
func NewBadgeSource(repository *groupdb.Repository) badge.Source { return repository }

// RegisterLifecycle validates and loads badge editor reference data at startup.
func RegisterLifecycle(lifecycle fx.Lifecycle, registry *badge.Registry) {
	lifecycle.Append(fx.Hook{OnStart: func(ctx context.Context) error { return registry.Refresh(ctx) }})
}

// RegisterRoomPolicy attaches warmed social-group decoration rights to active rooms.
func RegisterRoomPolicy(rooms *roomlive.Registry, groups *Service) {
	rooms.SetFurniturePolicy(groups.CanDecorate)
}

// HandlerDeps contains group packet adapter dependencies.
type HandlerDeps struct {
	fx.In

	// Identity manages creation and metadata.
	Identity *identity.Service
	// Membership manages social roles and preferences.
	Membership *membership.Service
	// Forum manages forum policy and durable content.
	Forum *forum.Service
	// Cursors stores ephemeral forum report context.
	Cursors *forum.Cursors
	// Delivery resolves authenticated connections.
	Delivery *groupruntime.Delivery
	// Cache stores warmed room and furniture generations.
	Cache *groupruntime.Cache
	// Translations localizes hotel-facing feedback.
	Translations i18n.Translator
	// Moderation accepts forum calls for help.
	Moderation *moderationcore.Service
	// Rooms stores active room presence for group-furniture context requests.
	Rooms *roomlive.Registry
}

// RegisterConnectionHandlers registers every consumed Nitro social-group packet.
func RegisterConnectionHandlers(handlers *realmconn.Handlers, dependencies HandlerDeps) {
	if handlers == nil || handlers.Inbound == nil {
		return
	}
	identity.RegisterHandlers(handlers.Inbound, identity.Handler{Identity: dependencies.Identity, Membership: dependencies.Membership, Delivery: dependencies.Delivery, Translations: dependencies.Translations})
	membership.RegisterHandlers(handlers.Inbound, membership.Handler{Membership: dependencies.Membership, Delivery: dependencies.Delivery, Translations: dependencies.Translations})
	forum.RegisterHandlers(handlers.Inbound, forum.Handler{Forum: dependencies.Forum, Cursors: dependencies.Cursors, Delivery: dependencies.Delivery, Translations: dependencies.Translations, Moderation: dependencies.Moderation})
	grouproom.RegisterHandlers(handlers.Inbound, grouproom.Handler{Cache: dependencies.Cache, Delivery: dependencies.Delivery, Rooms: dependencies.Rooms})
}
