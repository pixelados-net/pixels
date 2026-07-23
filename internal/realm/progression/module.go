// Package progression wires achievements, quests, talents, quizzes, and promotions.
package progression

import (
	"context"

	permissionservice "github.com/niflaot/pixels/internal/permission/service"
	realmconn "github.com/niflaot/pixels/internal/realm/connection"
	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
	currencyservice "github.com/niflaot/pixels/internal/realm/inventory/currency/service"
	moderationruntime "github.com/niflaot/pixels/internal/realm/moderation/runtime"
	playerachievement "github.com/niflaot/pixels/internal/realm/player/achievement"
	progressionachievement "github.com/niflaot/pixels/internal/realm/progression/achievement"
	compathandlers "github.com/niflaot/pixels/internal/realm/progression/compat/handlers"
	progressionconfig "github.com/niflaot/pixels/internal/realm/progression/config"
	progressiondb "github.com/niflaot/pixels/internal/realm/progression/database"
	progressionadmin "github.com/niflaot/pixels/internal/realm/progression/database/admin"
	progressionengine "github.com/niflaot/pixels/internal/realm/progression/engine"
	progressionobservability "github.com/niflaot/pixels/internal/realm/progression/observability"
	progressionpolicy "github.com/niflaot/pixels/internal/realm/progression/policy"
	progressionpoll "github.com/niflaot/pixels/internal/realm/progression/poll"
	pollhandlers "github.com/niflaot/pixels/internal/realm/progression/poll/handlers"
	progressionpromo "github.com/niflaot/pixels/internal/realm/progression/promo"
	promohandlers "github.com/niflaot/pixels/internal/realm/progression/promo/handlers"
	progressionquest "github.com/niflaot/pixels/internal/realm/progression/quest"
	questhandlers "github.com/niflaot/pixels/internal/realm/progression/quest/handlers"
	progressionquiz "github.com/niflaot/pixels/internal/realm/progression/quiz"
	quizhandlers "github.com/niflaot/pixels/internal/realm/progression/quiz/handlers"
	progressionrecord "github.com/niflaot/pixels/internal/realm/progression/record"
	progressiontalent "github.com/niflaot/pixels/internal/realm/progression/talent"
	talenthandlers "github.com/niflaot/pixels/internal/realm/progression/talent/handlers"
	progressiontrigger "github.com/niflaot/pixels/internal/realm/progression/trigger"
	roomentered "github.com/niflaot/pixels/internal/realm/room/access/events/entered"
	roombroadcast "github.com/niflaot/pixels/internal/realm/room/runtime/broadcast"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	tradecore "github.com/niflaot/pixels/internal/realm/trade/core"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	"github.com/niflaot/pixels/pkg/bus"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Module provides complete progression persistence, behavior, and packet routing.
var Module = fx.Module("realm-progression", fx.Provide(
	progressionconfig.Load, progressionobservability.New, progressiondb.New, progressionadmin.New, NewStore, NewAdminStore, NewCatalog,
	progressionengine.NewLiveProjector, NewBadgeManager, NewEngine,
	progressionquest.NewLiveProjector, NewQuestItems, NewQuest,
	progressiontalent.NewLiveProjector, NewTalentItems, NewTalent,
	NewQuiz, NewPromo, progressiontrigger.New, NewPoll,
), fx.Invoke(RegisterConnectionHandlers, RegisterLifecycle, RegisterGates, RegisterObservability, RegisterPollOffers, progressiontrigger.Register, progressiontrigger.RegisterPresence))

// NewStore exposes PostgreSQL persistence through the progression contract.
func NewStore(repository *progressiondb.Repository) progressionrecord.Store { return repository }

// NewAdminStore exposes PostgreSQL administration through the progression contract.
func NewAdminStore(repository *progressionadmin.Repository) progressionrecord.AdminStore {
	return repository
}

// NewCatalog creates the immutable cached progression catalog.
func NewCatalog(store progressionrecord.Store) *progressionengine.Catalog {
	return progressionengine.NewCatalog(store)
}

// NewBadgeManager adapts durable badge behavior to the progression engine boundary.
func NewBadgeManager(service *playerachievement.Service) progressionengine.BadgeManager {
	return service
}

// NewEngine creates the configured achievement trigger engine.
func NewEngine(config progressionconfig.Config, catalog *progressionengine.Catalog, store progressionrecord.Store, badges progressionengine.BadgeManager, currencies currencyservice.Granter, log *zap.Logger, projector *progressionengine.LiveProjector) *progressionengine.Service {
	return progressionengine.New(config, catalog, store, badges, currencies, log, projector)
}

// NewQuestItems adapts furniture grants to quest rewards.
func NewQuestItems(service *furnitureservice.Service) progressionquest.ItemGranter { return service }

// NewQuest creates durable quest behavior.
func NewQuest(config progressionconfig.Config, catalog *progressionengine.Catalog, store progressionrecord.Store, badges *playerachievement.Service, currencies currencyservice.Granter, items progressionquest.ItemGranter, projector *progressionquest.LiveProjector) *progressionquest.Service {
	return progressionquest.New(config, catalog, store, badges, currencies, items, projector)
}

// NewTalentItems adapts furniture grants to talent rewards.
func NewTalentItems(service *furnitureservice.Service) progressiontalent.ItemGranter { return service }

// NewTalent creates derived talent behavior.
func NewTalent(catalog *progressionengine.Catalog, store progressionrecord.Store, badges *playerachievement.Service, items progressiontalent.ItemGranter, permissions permissionservice.Manager, projector *progressiontalent.LiveProjector) *progressiontalent.Service {
	return progressiontalent.New(catalog, store, badges, items, permissions, projector)
}

// NewQuiz creates durable safety quiz behavior through the real progression engine.
func NewQuiz(catalog *progressionengine.Catalog, store progressionrecord.Store, engine *progressionengine.Service) *progressionquiz.Service {
	return progressionquiz.New(catalog, store, engine)
}

// NewPromo creates promotional badge claims through durable badge inventory.
func NewPromo(catalog *progressionengine.Catalog, store progressionrecord.Store, badges *playerachievement.Service) *progressionpromo.Service {
	return progressionpromo.New(catalog, store, badges)
}

// NewPoll creates the bounded word-poll room runtime.
func NewPoll(rooms *roomlive.Registry, connections *netconn.Registry, repository *progressiondb.Repository, badges *playerachievement.Service) *progressionpoll.Service {
	service := progressionpoll.New(rooms, func(ctx context.Context, room *roomlive.Room, packet codec.Packet, excludedPlayerID int64) error {
		return roombroadcast.RoomPacket(ctx, connections, room, packet, excludedPlayerID)
	})
	return service.WithDatabase(repository, badges)
}

// RegisterPollOffers sends unanswered DB poll offers on assigned room entry.
func RegisterPollOffers(lifecycle fx.Lifecycle, subscriber bus.Subscriber, polls *progressionpoll.Service, rooms *roomlive.Registry, connections *netconn.Registry) error {
	subscription, err := subscriber.Subscribe(roomentered.Name, bus.PriorityNormal, func(ctx context.Context, event bus.Event) error {
		payload, ok := event.Payload.(roomentered.Payload)
		if !ok {
			return nil
		}
		packet, offered, offerErr := polls.OfferForRoom(ctx, payload.PlayerID, payload.RoomID)
		if offerErr != nil || !offered {
			return offerErr
		}
		active, found := rooms.Find(payload.RoomID)
		if !found {
			return nil
		}
		occupant, found := active.Occupant(payload.PlayerID)
		if !found {
			return nil
		}
		connection, found := connections.Get(occupant.ConnectionKind, occupant.ConnectionID)
		if !found {
			return nil
		}
		return connection.Send(ctx, packet)
	})
	if err != nil {
		return err
	}
	lifecycle.Append(fx.Hook{OnStop: func(context.Context) error { subscription.Unsubscribe(); return nil }})
	return nil
}

// RegisterConnectionHandlers registers every real progression packet adapter.
func RegisterConnectionHandlers(handlers *realmconn.Handlers, catalog *progressionengine.Catalog, store progressionrecord.Store, quests *progressionquest.Service, talents *progressiontalent.Service, quizzes *progressionquiz.Service, polls *progressionpoll.Service, promos *progressionpromo.Service, badges *playerachievement.Service, bindings *binding.Registry) {
	if handlers == nil || handlers.Inbound == nil {
		return
	}
	progressionachievement.Register(handlers.Inbound, progressionachievement.Handler{Catalog: catalog, Store: store, Bindings: bindings})
	questhandlers.Register(handlers.Inbound, questhandlers.Handler{Service: quests, Bindings: bindings})
	talenthandlers.Register(handlers.Inbound, talenthandlers.Handler{Service: talents, Catalog: catalog, Store: store, Bindings: bindings})
	quizhandlers.Register(handlers.Inbound, quizhandlers.Handler{Service: quizzes, Bindings: bindings})
	pollhandlers.Register(handlers.Inbound, pollhandlers.Handler{Service: polls, Bindings: bindings})
	promohandlers.Register(handlers.Inbound, promohandlers.Handler{Service: promos, Badges: badges, Bindings: bindings})
	compathandlers.Register(handlers.Inbound, compathandlers.Handler{Catalog: catalog})
}

// RegisterLifecycle links circular capabilities and owns the write-behind worker.
func RegisterLifecycle(lifecycle fx.Lifecycle, engine *progressionengine.Service, quests *progressionquest.Service, talents *progressiontalent.Service, polls *progressionpoll.Service) {
	engine.SetQuestProgressor(quests)
	engine.SetTalentRecalculator(talents)
	lifecycle.Append(fx.Hook{OnStart: engine.Start, OnStop: engine.Stop})
	lifecycle.Append(fx.Hook{OnStart: polls.ReloadDatabase})
	lifecycle.Append(fx.Hook{OnStop: func(context.Context) error { polls.Close(); return nil }})
}

// RegisterGates connects optional progression policy to trade and guide realms.
func RegisterGates(config progressionconfig.Config, talents *progressiontalent.Service, trades *tradecore.Service, moderation *moderationruntime.Context) {
	trades.SetProgressionGate(config.TradeRequiresPerk, progressionpolicy.TradePerk)
	moderation.SetGuideEligibility(talents, config.GuideMinimumTrackLevel)
}

// RegisterObservability attaches lock-free metrics before serving progression work.
func RegisterObservability(metrics *progressionobservability.Metrics, catalog *progressionengine.Catalog, engine *progressionengine.Service, quests *progressionquest.Service, talents *progressiontalent.Service, promos *progressionpromo.Service) {
	catalog.SetMetrics(metrics)
	engine.SetMetrics(metrics)
	quests.SetMetrics(metrics)
	talents.SetMetrics(metrics)
	promos.SetMetrics(metrics)
}
