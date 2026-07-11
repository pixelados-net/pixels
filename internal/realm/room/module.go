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
	roomsettings "github.com/niflaot/pixels/internal/realm/room/settings"
	roomwordfilter "github.com/niflaot/pixels/internal/realm/room/wordfilter"
	wordrepo "github.com/niflaot/pixels/internal/realm/room/wordfilter/repository"
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
		NewConfigManager,
		NewRightsStore,
		NewRightsService,
		NewRightsManager,
		NewModerationStore,
		NewModerationService,
		NewModerationReader,
		NewModerationManager,
		NewAuditStore,
		NewAuditService,
		NewAuditManager,
		rightsbroadcast.New,
		moderationbroadcast.New,
		NewEntryService,
		NewSettingsAuthorizer,
		NewWordFilterStore,
		NewWordFilterService,
		NewWordFilterManager,
	),
	fx.Invoke(roomaudit.RegisterSubscriber),
	fx.Invoke(rightsbroadcast.Register),
	fx.Invoke(moderationbroadcast.Register),
	fx.Invoke(RegisterRuntimeCleanup),
	fx.Invoke(RegisterConnectionHandlers),
)

// NewSettingsAuthorizer creates shared room settings authorization.
func NewSettingsAuthorizer(permissions permissionservice.Checker) *roomsettings.Authorizer {
	return roomsettings.New(permissions, roomsettings.Nodes{
		OwnManage: SettingsOwnManage, AnyManage: SettingsAnyManage,
		OwnPolicyManage: ModerationOwnPolicyManage, AnyPolicyManage: ModerationAnyPolicyManage,
	})
}

// NewWordFilterStore creates room word filter persistence.
func NewWordFilterStore(pool *postgres.Pool) wordrepo.Store {
	return wordrepo.New(pool)
}

// NewWordFilterService creates room word filter behavior.
func NewWordFilterService(store wordrepo.Store, rooms service.Manager, authorize *roomsettings.Authorizer) *roomwordfilter.Service {
	return roomwordfilter.New(store, rooms, authorize)
}

// NewWordFilterManager exposes room word filter behavior through its contract.
func NewWordFilterManager(service *roomwordfilter.Service) roomwordfilter.Manager {
	return service
}

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

// NewRightsManager exposes room rights behavior through its contract.
func NewRightsManager(service *roomrights.Service) roomrights.Manager {
	return service
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

// NewModerationManager exposes room moderation behavior through its contract.
func NewModerationManager(moderation *roommoderation.Service) roommoderation.Manager {
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

// NewConfigManager exposes room settings persistence through its focused contract.
func NewConfigManager(roomService *service.Service) service.ConfigManager {
	return roomService
}
