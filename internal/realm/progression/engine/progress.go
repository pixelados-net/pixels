package engine

import (
	"context"
	"sync"
	"sync/atomic"

	currencyservice "github.com/niflaot/pixels/internal/realm/inventory/currency/service"
	progressionconfig "github.com/niflaot/pixels/internal/realm/progression/config"
	progressionobservability "github.com/niflaot/pixels/internal/realm/progression/observability"
	progressionrecord "github.com/niflaot/pixels/internal/realm/progression/record"
	"go.uber.org/zap"
)

// BadgeManager owns durable player badge inventory.
type BadgeManager interface {
	// GrantBadge grants one badge idempotently.
	GrantBadge(context.Context, int64, string, string) (bool, error)
	// ReplaceBadge replaces one badge while preserving its slot.
	ReplaceBadge(context.Context, int64, string, string, string) (bool, error)
	// RemoveBadge removes one badge regardless of equipped state.
	RemoveBadge(context.Context, int64, string) (bool, error)
}

// Projector publishes committed progression changes.
type Projector interface {
	// Project publishes one committed progression transition.
	Project(context.Context, Transition)
}

// QuestProgressor advances an active quest from a shared gameplay trigger.
type QuestProgressor interface {
	// ProgressTrigger advances a matching active quest.
	ProgressTrigger(context.Context, int64, string, string, int64) error
}

// TalentRecalculator advances derived talent tracks after achievement changes.
type TalentRecalculator interface {
	// Recalculate evaluates tracks affected by one achievement definition.
	Recalculate(context.Context, int64, int64) error
}

// Transition describes one committed achievement transition.
type Transition struct {
	// PlayerID identifies the affected player.
	PlayerID int64
	// Definition identifies the achievement group.
	Definition progressionrecord.AchievementDefinition
	// Mutation stores before, after, and crossed levels.
	Mutation progressionrecord.ProgressMutation
	// Score stores the new durable score.
	Score int32
}

// Service executes queued and direct progression mutations.
type Service struct {
	// config stores runtime policy.
	config progressionconfig.Config
	// catalog owns immutable definitions.
	catalog *Catalog
	// store persists player state.
	store progressionrecord.Store
	// badges owns badge inventory.
	badges BadgeManager
	// currencies owns wallet rewards.
	currencies currencyservice.Granter
	// projector publishes post-commit packets.
	projector Projector
	// quests advances the player's active quest.
	quests QuestProgressor
	// talents advances affected derived tracks.
	talents TalentRecalculator
	// log records asynchronous failures.
	log *zap.Logger
	// queue accepts non-blocking gameplay triggers.
	queue chan trigger
	// stop terminates the worker.
	stop chan struct{}
	// done confirms worker termination.
	done chan struct{}
	// started protects lifecycle idempotence.
	started atomic.Bool
	// mutex protects pending deltas.
	mutex sync.Mutex
	// pending groups deferred trigger deltas.
	pending map[triggerKey]int64
	// forecast stores queued achievement deltas by player and trigger.
	forecast map[playerTrigger]int64
	// known stores hydrated or freshly committed achievement progress.
	known map[playerAchievement]int64
	// hydrated identifies players whose absent progress rows are known zeros.
	hydrated map[int64]bool
	// metrics stores process-wide progression telemetry.
	metrics *progressionobservability.Metrics
}

// SetMetrics attaches process-wide telemetry before serving triggers.
func (service *Service) SetMetrics(metrics *progressionobservability.Metrics) {
	service.metrics = metrics
}

// trigger stores one queued gameplay signal.
type trigger struct {
	// playerID identifies the affected player.
	playerID int64
	// key identifies the trigger.
	key string
	// data stores optional goal-specific trigger metadata.
	data string
	// amount stores the positive delta.
	amount int64
	// daily enables UTC-day idempotence.
	daily bool
}

// triggerKey groups deferred deltas.
type triggerKey struct {
	// playerID identifies the affected player.
	playerID int64
	// key identifies the trigger.
	key string
	// data keeps distinct goal metadata in separate batches.
	data string
	// daily preserves daily idempotence semantics.
	daily bool
}

// New creates a progression engine.
func New(config progressionconfig.Config, catalog *Catalog, store progressionrecord.Store, badges BadgeManager, currencies currencyservice.Granter, log *zap.Logger, projectors ...Projector) *Service {
	if log == nil {
		log = zap.NewNop()
	}
	service := &Service{config: config, catalog: catalog, store: store, badges: badges, currencies: currencies, log: log, queue: make(chan trigger, 1024), stop: make(chan struct{}), done: make(chan struct{}), pending: make(map[triggerKey]int64), forecast: make(map[playerTrigger]int64), known: make(map[playerAchievement]int64), hydrated: make(map[int64]bool)}
	if len(projectors) > 0 {
		service.projector = projectors[0]
	}
	return service
}

// Start loads the catalog and starts deferred persistence.
func (service *Service) Start(ctx context.Context) error {
	if service == nil || !service.config.Enabled || !service.started.CompareAndSwap(false, true) {
		return nil
	}
	if err := service.validateDependencies(); err != nil {
		service.started.Store(false)
		return err
	}
	if err := service.catalog.Reload(ctx); err != nil {
		service.started.Store(false)
		return err
	}
	go service.run()
	return nil
}

// Stop flushes all pending progress and stops deferred persistence.
func (service *Service) Stop(ctx context.Context) error {
	if service == nil || !service.started.CompareAndSwap(true, false) {
		return nil
	}
	close(service.stop)
	select {
	case <-service.done:
		return service.flush(context.Background(), 0)
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Progress queues one gameplay trigger without blocking its publisher.
func (service *Service) Progress(ctx context.Context, playerID int64, key string, amount int64) error {
	return service.progress(ctx, playerID, key, "", amount, false)
}

// ProgressData queues one gameplay trigger with goal-specific metadata.
func (service *Service) ProgressData(ctx context.Context, playerID int64, key string, data string, amount int64) error {
	return service.progress(ctx, playerID, key, data, amount, false)
}

// ProgressDaily queues one idempotent UTC-day trigger.
func (service *Service) ProgressDaily(ctx context.Context, playerID int64, key string, amount int64) error {
	return service.progress(ctx, playerID, key, "", amount, true)
}

// progress validates and queues one trigger.
func (service *Service) progress(ctx context.Context, playerID int64, key string, data string, amount int64, daily bool) error {
	if service == nil || !service.config.Enabled || playerID <= 0 || amount <= 0 {
		return nil
	}
	if len(service.catalog.Achievements(key)) == 0 && len(service.catalog.Quests(key)) == 0 {
		return nil
	}
	value := trigger{playerID: playerID, key: key, data: data, amount: amount, daily: daily}
	service.metrics.RecordTrigger(key)
	select {
	case service.queue <- value:
		service.metrics.AddQueue(1)
		return nil
	default:
		service.add(value)
		if err := service.FlushPlayer(ctx, playerID); err != nil {
			return err
		}
		return nil
	}
}

// ProgressNow applies one trigger synchronously without write-behind.
func (service *Service) ProgressNow(ctx context.Context, playerID int64, key string, amount int64, daily bool) error {
	return service.ProgressNowData(ctx, playerID, key, "", amount, daily)
}

// ProgressNowData applies one metadata-bearing trigger synchronously.
func (service *Service) ProgressNowData(ctx context.Context, playerID int64, key string, data string, amount int64, daily bool) error {
	if service == nil || !service.config.Enabled || playerID <= 0 || amount <= 0 {
		return nil
	}
	for _, definition := range service.catalog.Achievements(key) {
		if err := service.apply(ctx, playerID, *definition, amount, daily); err != nil {
			return err
		}
	}
	if service.quests != nil && len(service.catalog.Quests(key)) > 0 {
		return service.quests.ProgressTrigger(ctx, playerID, key, data, amount)
	}
	return nil
}
