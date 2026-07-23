package chat

import (
	permissionservice "github.com/niflaot/pixels/internal/permission/service"
	"github.com/niflaot/pixels/internal/realm/chat/bubble"
	bubblerepo "github.com/niflaot/pixels/internal/realm/chat/bubble/repository"
	chatconfig "github.com/niflaot/pixels/internal/realm/chat/config"
	chatfilter "github.com/niflaot/pixels/internal/realm/chat/filter"
	filterrepo "github.com/niflaot/pixels/internal/realm/chat/filter/repository"
	"github.com/niflaot/pixels/internal/realm/chat/history"
	historyrepo "github.com/niflaot/pixels/internal/realm/chat/history/repository"
	chatsend "github.com/niflaot/pixels/internal/realm/chat/send"
	playeridentity "github.com/niflaot/pixels/internal/realm/player/identity"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	playerrepo "github.com/niflaot/pixels/internal/realm/player/repository"
	roomrealm "github.com/niflaot/pixels/internal/realm/room"
	roomwordfilter "github.com/niflaot/pixels/internal/realm/room/control/wordfilter"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	wiredbridge "github.com/niflaot/pixels/internal/realm/room/world/wired/bridge"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	netconn "github.com/niflaot/pixels/networking/connection"
	"github.com/niflaot/pixels/pkg/bus"
	"github.com/niflaot/pixels/pkg/i18n"
	"github.com/niflaot/pixels/pkg/postgres"
	"github.com/niflaot/pixels/pkg/redis"
	"go.uber.org/fx"
)

// Module provides chat persistence, validation, and protocol handlers.
var Module = fx.Module(
	"realm-chat",
	fx.Provide(
		NewFilterStore,
		chatfilter.New,
		NewIdentityWordFilter,
		NewBubbleStore,
		NewBubbleProfileStore,
		NewBubblePermissions,
		NewBubbleService,
		NewSendService,
		NewHistoryStore,
		history.NewWriter,
		history.NewService,
	),
	fx.Invoke(RefreshGlobalFilter),
	fx.Invoke(history.Register),
	fx.Invoke(RegisterConnectionHandlers),
	fx.Invoke(RegisterWiredSpeech),
)

// NewIdentityWordFilter exposes the immutable hotel dictionary to username policy.
func NewIdentityWordFilter(service *chatfilter.Service) playeridentity.WordFilter { return service }

// RegisterWiredSpeech installs filtered player talk and shout interception.
func RegisterWiredSpeech(service *chatsend.Service, bridge *wiredbridge.SpeechBridge) {
	service.SetSpeechInterceptor(bridge)
}

// NewFilterStore creates global chat filter persistence.
func NewFilterStore(pool *postgres.Pool) filterrepo.Store { return filterrepo.New(pool) }

// NewBubbleStore creates bubble unlock persistence.
func NewBubbleStore(pool *postgres.Pool) bubblerepo.Store { return bubblerepo.New(pool) }

// NewHistoryStore creates partitioned chat history persistence.
func NewHistoryStore(pool *postgres.Pool) historyrepo.Store { return historyrepo.New(pool) }

// NewBubbleProfileStore exposes player profile bubble persistence.
func NewBubbleProfileStore(store playerrepo.Store) bubble.ProfileStore { return store }

// NewBubblePermissions exposes the focused bubble permission contract.
func NewBubblePermissions(manager permissionservice.Manager) bubble.PermissionReader { return manager }

// NewBubbleService creates bubble selection behavior.
func NewBubbleService(store bubblerepo.Store, profiles bubble.ProfileStore, permissions bubble.PermissionReader, players *playerlive.Registry) *bubble.Service {
	return bubble.New(store, profiles, permissions, players, BubbleAny)
}

// NewSendService composes the room chat hot path.
func NewSendService(config chatconfig.Config, players *playerlive.Registry, bindings *binding.Registry, rooms *roomlive.Registry, connections *netconn.Registry, permissions permissionservice.Checker, counter *redis.Client, globalFilter *chatfilter.Service, roomFilter roomwordfilter.Manager, events bus.Publisher, translations i18n.Translator) *chatsend.Service {
	return chatsend.New(config, players, bindings, rooms, connections, permissions, counter, globalFilter, roomFilter, events, translations, chatsend.Nodes{
		FloodImmune: FloodImmune, LengthUnlimited: LengthUnlimited, FilterImmune: FilterImmune,
		WhisperObserveAny: WhisperObserveAny, ModerationOwnMute: roomrealm.ModerationOwnMute,
		ModerationAnyMute: roomrealm.ModerationAnyMute,
	})
}

// RefreshGlobalFilter warms the immutable global dictionary before serving chat.
func RefreshGlobalFilter(lifecycle fx.Lifecycle, service *chatfilter.Service) {
	lifecycle.Append(fx.Hook{OnStart: service.Refresh})
}
