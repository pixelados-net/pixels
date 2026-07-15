// Package room contains room realm persistence and runtime wiring.
package room

import (
	permissionservice "github.com/niflaot/pixels/internal/permission/service"
	roomentry "github.com/niflaot/pixels/internal/realm/room/access/entry"
	roomaudit "github.com/niflaot/pixels/internal/realm/room/control/audit"
	roomfloorplan "github.com/niflaot/pixels/internal/realm/room/control/floorplan"
	roommoderation "github.com/niflaot/pixels/internal/realm/room/control/moderation"
	moderationbroadcast "github.com/niflaot/pixels/internal/realm/room/control/moderation/broadcast"
	roomrights "github.com/niflaot/pixels/internal/realm/room/control/rights"
	rightsbroadcast "github.com/niflaot/pixels/internal/realm/room/control/rights/broadcast"
	roomsettings "github.com/niflaot/pixels/internal/realm/room/control/settings"
	roomvotes "github.com/niflaot/pixels/internal/realm/room/control/votes"
	roomwordfilter "github.com/niflaot/pixels/internal/realm/room/control/wordfilter"
	auditrepo "github.com/niflaot/pixels/internal/realm/room/database/audit"
	decorationrepo "github.com/niflaot/pixels/internal/realm/room/database/decoration"
	layoutrepo "github.com/niflaot/pixels/internal/realm/room/database/layout"
	moderationrepo "github.com/niflaot/pixels/internal/realm/room/database/moderation"
	"github.com/niflaot/pixels/internal/realm/room/database/record"
	rightsrepo "github.com/niflaot/pixels/internal/realm/room/database/rights"
	votesrepo "github.com/niflaot/pixels/internal/realm/room/database/votes"
	wordrepo "github.com/niflaot/pixels/internal/realm/room/database/wordfilter"
	roomdecoration "github.com/niflaot/pixels/internal/realm/room/decoration"
	roombundle "github.com/niflaot/pixels/internal/realm/room/record/bundle"
	"github.com/niflaot/pixels/internal/realm/room/record/service"
	roomaction "github.com/niflaot/pixels/internal/realm/room/world/action"
	"github.com/niflaot/pixels/internal/realm/room/world/layout"
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
		decorationrepo.New,
		NewDecorationStore,
		roomdecoration.New,
		roomfloorplan.LoadConfig,
		layout.NewService,
		service.New,
		NewBundleStore,
		roombundle.LoadConfig,
		roombundle.New,
		NewLiveRegistry,
		NewLayoutManager,
		NewRoomLayoutManager,
		NewManager,
		NewConfigManager,
		NewBundleManager,
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
		NewFloorplanAuthorizer,
		NewWordFilterStore,
		NewWordFilterService,
		NewWordFilterManager,
		NewVoteStore,
		NewVoteRoomFinder,
		roomvotes.New,
		NewVoteManager,
		roomaction.LoadConfig,
		roomaction.New,
		roomaction.NewScheduler,
	),
	fx.Invoke(roomaudit.RegisterSubscriber),
	fx.Invoke(rightsbroadcast.Register),
	fx.Invoke(moderationbroadcast.Register),
	fx.Invoke(RegisterRuntimeCleanup),
	fx.Invoke(RegisterConnectionHandlers),
	fx.Invoke(roomaction.RegisterScheduler),
)

// NewDecorationStore exposes PostgreSQL room decoration persistence through its domain contract.
func NewDecorationStore(repository *decorationrepo.Repository) roomdecoration.Store {
	return repository
}

// NewVoteStore creates room vote persistence.
func NewVoteStore(pool *postgres.Pool) roomvotes.Store {
	return votesrepo.New(pool)
}

// NewVoteRoomFinder exposes room reads to vote behavior.
func NewVoteRoomFinder(service *service.Service) roomvotes.RoomFinder {
	return service
}

// NewVoteManager exposes room vote behavior through its contract.
func NewVoteManager(service *roomvotes.Service) roomvotes.Manager {
	return service
}

// NewSettingsAuthorizer creates shared room settings authorization.
func NewSettingsAuthorizer(permissions permissionservice.Checker) *roomsettings.Authorizer {
	return roomsettings.New(permissions, roomsettings.Nodes{
		OwnManage: SettingsOwnManage, AnyManage: SettingsAnyManage,
		OwnPolicyManage: ModerationOwnPolicyManage, AnyPolicyManage: ModerationAnyPolicyManage,
	})
}

// NewFloorplanAuthorizer creates shared floor plan authorization.
func NewFloorplanAuthorizer(permissions permissionservice.Checker, rights *roomrights.Service) *roomfloorplan.Authorizer {
	return roomfloorplan.NewAuthorizer(permissions, rights, roomfloorplan.Nodes{
		OwnEdit: FloorplanOwnEdit, AnyEdit: FloorplanAnyEdit,
	})
}

// NewWordFilterStore creates room word filter persistence.
func NewWordFilterStore(pool *postgres.Pool) roomwordfilter.Store {
	return wordrepo.New(pool)
}

// NewWordFilterService creates room word filter behavior.
func NewWordFilterService(store roomwordfilter.Store, rooms service.Manager, authorize *roomsettings.Authorizer) *roomwordfilter.Service {
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
func NewRightsStore(pool *postgres.Pool) roomrights.Store {
	return rightsrepo.New(pool)
}

// NewRightsService creates room rights behavior.
func NewRightsService(store roomrights.Store, rooms *service.Service, permissions permissionservice.Checker, events bus.Publisher) *roomrights.Service {
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
func NewModerationStore(pool *postgres.Pool) roommoderation.Store {
	return moderationrepo.New(pool)
}

// NewModerationService creates room moderation behavior.
func NewModerationService(config roommoderation.Config, store roommoderation.Store, rooms *service.Service, rights *roomrights.Service, permissions permissionservice.Checker, events bus.Publisher) *roommoderation.Service {
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
func NewAuditStore(pool *postgres.Pool) roomaudit.Store {
	return auditrepo.New(pool)
}

// NewAuditService creates room audit query behavior.
func NewAuditService(store roomaudit.Store) *roomaudit.Service {
	return roomaudit.New(store)
}

// NewAuditManager exposes room audit queries.
func NewAuditManager(audit *roomaudit.Service) roomaudit.Manager {
	return audit
}

// NewLayoutStore creates the room layout persistence store.
func NewLayoutStore(pool *postgres.Pool) layout.Store {
	return layoutrepo.NewRepository(pool)
}

// NewStore creates the room persistence store.
func NewStore(pool *postgres.Pool) service.Store {
	return repository.New(pool)
}

// NewLayoutManager exposes the room layout management contract.
func NewLayoutManager(service *layout.Service) layout.Manager {
	return service
}

// NewRoomLayoutManager exposes room-owned layout management.
func NewRoomLayoutManager(service *layout.Service) layout.RoomManager {
	return service
}

// NewRoomService creates room persistence behavior with bundle dependencies.
func NewBundleStore(pool *postgres.Pool) roombundle.Store {
	return repository.New(pool)
}

// NewBundleManager exposes room bundle administration and purchase behavior.
func NewBundleManager(roomBundles *roombundle.Service) roombundle.Manager {
	return roomBundles
}

// NewManager exposes the room management contract.
func NewManager(roomService *service.Service) service.Manager {
	return roomService
}

// NewConfigManager exposes room settings persistence through its focused contract.
func NewConfigManager(roomService *service.Service) service.ConfigManager {
	return roomService
}
