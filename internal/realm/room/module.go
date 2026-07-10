// Package room contains room realm persistence and runtime wiring.
package room

import (
	permissionservice "github.com/niflaot/pixels/internal/permission/service"
	roomaudit "github.com/niflaot/pixels/internal/realm/room/audit"
	auditrepo "github.com/niflaot/pixels/internal/realm/room/audit/repository"
	roomentry "github.com/niflaot/pixels/internal/realm/room/entry"
	"github.com/niflaot/pixels/internal/realm/room/layout"
	roommoderation "github.com/niflaot/pixels/internal/realm/room/moderation"
	moderationbroadcast "github.com/niflaot/pixels/internal/realm/room/moderation/broadcast"
	moderationrepo "github.com/niflaot/pixels/internal/realm/room/moderation/repository"
	"github.com/niflaot/pixels/internal/realm/room/repository"
	roomrights "github.com/niflaot/pixels/internal/realm/room/rights"
	rightsbroadcast "github.com/niflaot/pixels/internal/realm/room/rights/broadcast"
	rightsrepo "github.com/niflaot/pixels/internal/realm/room/rights/repository"
	"github.com/niflaot/pixels/internal/realm/room/service"
	"github.com/niflaot/pixels/pkg/bus"
	"github.com/niflaot/pixels/pkg/i18n"
	"github.com/niflaot/pixels/pkg/postgres"
	"github.com/niflaot/pixels/pkg/redis"
	"go.uber.org/fx"
)

// Module provides room realm persistence state.
var Module = fx.Module(
	"realm-room",
	fx.Provide(
		NewLayoutStore,
		NewStore,
		layout.NewService,
		service.New,
		NewLiveRegistry,
		NewLayoutManager,
		NewManager,
		NewRightsStore,
		NewRightsService,
		NewModerationStore,
		NewModerationService,
		NewModerationReader,
		NewAuditStore,
		NewAuditService,
		NewAuditManager,
		rightsbroadcast.New,
		moderationbroadcast.New,
		NewEntryService,
	),
	fx.Invoke(roomaudit.RegisterSubscriber),
	fx.Invoke(rightsbroadcast.Register),
	fx.Invoke(moderationbroadcast.Register),
	fx.Invoke(RegisterRuntimeCleanup),
	fx.Invoke(RegisterConnectionHandlers),
)

// NewEntryService creates closed-room entry behavior.
func NewEntryService(config roomentry.Config, redisClient *redis.Client, permissions permissionservice.Checker, translations i18n.Translator, rights *roomrights.Service, moderation *roommoderation.Service) *roomentry.Service {
	service := roomentry.New(config, redisClient, permissions, translations, roomentry.Nodes{
		EnterAny: EnterAny, EnterFull: EnterFull, AnswerAnyDoorbell: AnswerAnyDoorbell,
	})

	return service.WithRights(rights).WithBans(moderation)
}

// NewRightsStore creates room rights persistence.
func NewRightsStore(pool *postgres.Pool) rightsrepo.Store {
	return rightsrepo.New(pool)
}

// NewRightsService creates room rights behavior.
func NewRightsService(store rightsrepo.Store, rooms *service.Service, permissions permissionservice.Checker, events bus.Publisher) *roomrights.Service {
	return roomrights.New(store, rooms, permissions, events, roomrights.Nodes{
		OwnGrant: RightsOwnGrant, OwnRevoke: RightsOwnRevoke,
		AnyGrant: RightsAnyGrant, AnyRevoke: RightsAnyRevoke,
	})
}

// NewModerationStore creates room moderation persistence.
func NewModerationStore(pool *postgres.Pool) moderationrepo.Store {
	return moderationrepo.New(pool)
}

// NewModerationService creates room moderation behavior.
func NewModerationService(config roommoderation.Config, store moderationrepo.Store, rooms *service.Service, rights *roomrights.Service, permissions permissionservice.Checker, events bus.Publisher) *roommoderation.Service {
	return roommoderation.New(config, store, rooms, rights, permissions, events, roommoderation.Nodes{
		OwnKick: ModerationOwnKick, OwnMute: ModerationOwnMute, OwnBan: ModerationOwnBan,
		AnyKick: ModerationAnyKick, AnyMute: ModerationAnyMute, AnyBan: ModerationAnyBan,
		Unkickable: Unkickable,
	})
}

// NewModerationReader exposes room moderation reads.
func NewModerationReader(moderation *roommoderation.Service) roommoderation.Reader {
	return moderation
}

// NewAuditStore creates room rights and moderation audit persistence.
func NewAuditStore(pool *postgres.Pool) auditrepo.Store {
	return auditrepo.New(pool)
}

// NewAuditService creates room audit query behavior.
func NewAuditService(store auditrepo.Store) *roomaudit.Service {
	return roomaudit.New(store)
}

// NewAuditManager exposes room audit queries.
func NewAuditManager(audit *roomaudit.Service) roomaudit.Manager {
	return audit
}

// NewLayoutStore creates the room layout persistence store.
func NewLayoutStore(pool *postgres.Pool) layout.Store {
	return layout.NewRepository(pool)
}

// NewStore creates the room persistence store.
func NewStore(pool *postgres.Pool) repository.Store {
	return repository.New(pool)
}

// NewLayoutManager exposes the room layout management contract.
func NewLayoutManager(service *layout.Service) layout.Manager {
	return service
}

// NewManager exposes the room management contract.
func NewManager(roomService *service.Service) service.Manager {
	return roomService
}
