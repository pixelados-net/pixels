package messenger

import (
	permissionservice "github.com/niflaot/pixels/internal/permission/service"
	chatfilter "github.com/niflaot/pixels/internal/realm/chat/filter"
	realmconn "github.com/niflaot/pixels/internal/realm/connection"
	messengercore "github.com/niflaot/pixels/internal/realm/messenger/core"
	messengerdatabase "github.com/niflaot/pixels/internal/realm/messenger/database"
	"github.com/niflaot/pixels/internal/realm/messenger/friend"
	"github.com/niflaot/pixels/internal/realm/messenger/profile"
	messengerrecord "github.com/niflaot/pixels/internal/realm/messenger/record"
	"github.com/niflaot/pixels/internal/realm/messenger/runtime/chatlog"
	"github.com/niflaot/pixels/internal/realm/messenger/runtime/delivery"
	"github.com/niflaot/pixels/internal/realm/messenger/session"
	"github.com/niflaot/pixels/internal/realm/messenger/social"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	playerservice "github.com/niflaot/pixels/internal/realm/player/service"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/pkg/bus"
	"github.com/niflaot/pixels/pkg/i18n"
	"github.com/niflaot/pixels/pkg/postgres"
	"github.com/niflaot/pixels/pkg/redis"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Module provides messenger persistence, behavior, and packet routing.
var Module = fx.Module(
	"realm-messenger",
	fx.Provide(NewStore, NewPrivateChatWriter, NewService, delivery.New, profile.NewPresence, NewProfileBroadcaster),
	fx.Invoke(RegisterConnectionHandlers, profile.RegisterPresence, profile.RegisterRelationships, chatlog.RegisterLifecycle),
)

// NewProfileBroadcaster creates targeted live profile relationship projection.
func NewProfileBroadcaster(messenger *messengercore.Service, sender *delivery.Sender, log *zap.Logger) *profile.RelationshipBroadcaster {
	return profile.NewRelationships(messenger, sender, log)
}

// NewStore creates messenger persistence behavior.
func NewStore(pool *postgres.Pool) messengerrecord.Store {
	return messengerdatabase.New(pool)
}

// NewService creates configured messenger behavior.
func NewService(config Config, store messengerrecord.Store, players playerservice.Manager, livePlayers *playerlive.Registry, rooms *roomlive.Registry, permissions permissionservice.Checker, redisClient *redis.Client, filter *chatfilter.Service, messageLog *chatlog.Writer) *messengercore.Service {
	config = config.Normalize()
	return messengercore.New(messengercore.Options{
		MaxFriends: config.MaxFriends, MaxFriendsClub: config.MaxFriendsClub,
		MaxSearchResults: config.MaxSearchResults, SearchCacheTTL: config.SearchCacheTTL,
		SearchThrottle: config.SearchThrottle, ChatThrottle: config.ChatThrottle,
		ChatFilterEnabled: config.ChatFilterEnabled, ChatLogEnabled: config.ChatLogEnabled,
	}, store, players, livePlayers, rooms, permissions, redisClient, filter, messengercore.Nodes{FriendsUnlimited: FriendsUnlimited, FollowAny: FollowAny}, messageLog)
}

// NewPrivateChatWriter creates optional asynchronous private-message persistence.
func NewPrivateChatWriter(config Config, store messengerrecord.Store, log *zap.Logger) *chatlog.Writer {
	return chatlog.New(chatlog.Config{Enabled: config.ChatLogEnabled}, store, log)
}

// HandlerDeps contains messenger packet handler dependencies.
type HandlerDeps struct {
	fx.In

	// Messenger stores messenger behavior.
	Messenger *messengercore.Service
	// Delivery sends packets through authenticated bindings.
	Delivery *delivery.Sender
	// Events publishes completed messenger actions.
	Events bus.Publisher
	// Translations localizes hotel-facing feedback.
	Translations i18n.Translator
	// Log records command dispatch and unexpected failures.
	Log *zap.Logger
}

// RegisterConnectionHandlers registers every Nitro messenger inbound packet.
func RegisterConnectionHandlers(handlers *realmconn.Handlers, dependencies HandlerDeps) {
	if handlers == nil || handlers.Inbound == nil {
		return
	}
	session.RegisterHandlers(handlers.Inbound, session.Handler{Messenger: dependencies.Messenger, Delivery: dependencies.Delivery}, dependencies.Log)
	friend.RegisterHandlers(handlers.Inbound, friend.Handler{Messenger: dependencies.Messenger, Delivery: dependencies.Delivery, Events: dependencies.Events}, dependencies.Log)
	social.RegisterHandlers(handlers.Inbound, social.Handler{Messenger: dependencies.Messenger, Delivery: dependencies.Delivery, Events: dependencies.Events, Translations: dependencies.Translations}, dependencies.Log)
	session.RegisterPrivacyHandlers(handlers.Inbound, session.PrivacyHandler{Messenger: dependencies.Messenger, Delivery: dependencies.Delivery}, dependencies.Log)
}
