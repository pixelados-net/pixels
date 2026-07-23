// Package runtime owns room-local pet controllers and projections.
package runtime

import (
	"context"
	"math/rand/v2"
	"sync"
	"sync/atomic"
	"time"

	petobservability "github.com/niflaot/pixels/internal/realm/pet/observability"
	petpolicy "github.com/niflaot/pixels/internal/realm/pet/policy"
	petrecord "github.com/niflaot/pixels/internal/realm/pet/record"
	petreference "github.com/niflaot/pixels/internal/realm/pet/reference"
	petwork "github.com/niflaot/pixels/internal/realm/pet/runtime/work"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	"github.com/niflaot/pixels/internal/realm/room/runtime/broadcast"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	roomprojection "github.com/niflaot/pixels/internal/realm/room/runtime/projection"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	"github.com/niflaot/pixels/networking/connection"
	outstatus "github.com/niflaot/pixels/networking/outbound/room/entities/status"
	"github.com/niflaot/pixels/pkg/bus"
	"github.com/niflaot/pixels/pkg/i18n"
	"go.uber.org/zap"
)

// EntityKey returns the collision-free world key for one pet.
func EntityKey(petID int64) int64 { return entityBase + petID }

// Source supplies deterministic random values in tests.
type Source interface {
	// Uint64 returns one pseudo-random value.
	Uint64() uint64
}

// Clock supplies deterministic wall time for lifecycle rules.
type Clock interface {
	// Now returns the current wall time.
	Now() time.Time
}

// NeedConsumer applies one autonomous furniture need away from the room tick.
type NeedConsumer interface {
	// ConsumeNeed applies one nearby food, drink, toy, or nest to a visible pet.
	ConsumeNeed(context.Context, int64, int64, int64) error
}

// SpeechInterceptor feeds localized pet speech into room automation.
type SpeechInterceptor interface {
	// InterceptPet examines one pet message before normal delivery.
	InterceptPet(context.Context, int64, int64, string) (bool, error)
}

// InventoryInvalidator drops warmed owner snapshots after committed mutations.
type InventoryInvalidator interface {
	// Invalidate removes one owner inventory generation.
	Invalidate(int64)
}

// globalSource uses Go's concurrency-safe random generator.
type globalSource struct{}

// Uint64 returns one pseudo-random value.
func (globalSource) Uint64() uint64 { return rand.Uint64() }

// systemClock supplies production wall time.
type systemClock struct{}

// Now returns the current wall time.
func (systemClock) Now() time.Time { return time.Now() }

// activePet stores one mutable room-local controller.
type activePet struct {
	// mutex protects fields touched by handlers and owner cycles.
	mutex sync.Mutex
	// record stores the latest durable snapshot.
	record petrecord.Pet
	// stationary prevents plant-controlled locomotion.
	stationary bool
	// plantStage stores the last projected renderer lifecycle stage.
	plantStage int8
	// nextDue stores the next autonomous decision deadline.
	nextDue time.Time
	// lastFlush stores the last position persistence cadence.
	lastFlush time.Time
	// lastPoint stores the last queued persistent tile.
	lastPoint grid.Point
	// actionGeneration invalidates stale scheduled completions.
	actionGeneration uint64
	// selectedBy stores the last player selecting this pet.
	selectedBy int64
	// followingPlayerID stores the current follow target.
	followingPlayerID int64
	// stay reports autonomous movement suppression.
	stay bool
	// cooldowns stores fixed command cooldown deadlines without a map allocation.
	cooldowns [47]time.Time
	// riderPlayerID stores the currently mounted player.
	riderPlayerID int64
	// needPending prevents duplicate autonomous product transactions.
	needPending bool
	// commandNeed stores one contextual command awaiting arrival or consumption.
	commandNeed commandNeedState
	// silent reports persistent vocal suppression.
	silent bool
	// nextVocal stores the earliest autonomous speech deadline.
	nextVocal time.Time
}

// roomState stores one active room pet generation.
type roomState struct {
	// pets stores controllers by durable pet identifier.
	pets map[int64]*activePet
	// snapshot stores an immutable zero-allocation cycle view.
	snapshot atomic.Pointer[petSnapshot]
	// ready closes after the initial durable room load finishes.
	ready chan struct{}
	// loadErr stores an initial load failure for concurrent waiters.
	loadErr error
}

// petSnapshot stores stable controller pointers.
type petSnapshot struct {
	// pets stores immutable pointer membership.
	pets []*activePet
}

// Service coordinates durable and active pet state.
type Service struct {
	// config stores normalized behavior limits.
	config petpolicy.Config
	// store persists pet aggregates.
	store petrecord.Store
	// references resolves immutable species and command data.
	references petreference.Reader
	// rooms stores active room worlds.
	rooms *roomlive.Registry
	// players resolves authenticated player peers.
	players *playerlive.Registry
	// connections sends Nitro projections.
	connections *connection.Registry
	// events publishes pet lifecycle events.
	events bus.Publisher
	// translations resolves localized pet vocals.
	translations i18n.Translator
	// speech feeds pet vocals into WIRED before delivery.
	speech SpeechInterceptor
	// needs applies autonomous products through the furniture transaction boundary.
	needs NeedConsumer
	// inventories invalidates warmed owner lists after committed mutations.
	inventories InventoryInvalidator
	// source supplies deterministic decisions.
	source Source
	// clock supplies deterministic lifecycle time.
	clock Clock
	// log records deferred persistence failures.
	log *zap.Logger
	// mutex protects room state generations.
	mutex sync.RWMutex
	// active stores loaded state by room identifier.
	active map[int64]*roomState
	// work runs persistence away from owner loops.
	work *petwork.Pool
	// metrics stores lock-free low-cardinality telemetry.
	metrics *petobservability.Metrics
}

// New creates pet runtime behavior.
func New(config petpolicy.Config, store petrecord.Store, references petreference.Reader, rooms *roomlive.Registry, players *playerlive.Registry, connections *connection.Registry, events bus.Publisher, translations i18n.Translator, speech SpeechInterceptor, log *zap.Logger, metrics *petobservability.Metrics) *Service {
	return &Service{config: config.Normalize(), store: store, references: references, rooms: rooms, players: players, connections: connections, events: events, translations: translations, speech: speech, source: globalSource{}, clock: systemClock{}, log: log, active: make(map[int64]*roomState), work: petwork.New(512, 2, log), metrics: metrics}
}

// SetSource replaces autonomous randomness for deterministic tests.
func (service *Service) SetSource(source Source) {
	if source != nil {
		service.source = source
	}
}

// SetClock replaces wall time for deterministic tests.
func (service *Service) SetClock(clock Clock) {
	if clock != nil {
		service.clock = clock
	}
}

// Now returns the configured lifecycle time.
func (service *Service) Now() time.Time {
	if service.clock == nil {
		return time.Now()
	}
	return service.clock.Now()
}

// SetNeedConsumer installs autonomous product behavior before room cycles start.
func (service *Service) SetNeedConsumer(consumer NeedConsumer) { service.needs = consumer }

// SetInventoryInvalidator installs the owner inventory cache boundary.
func (service *Service) SetInventoryInvalidator(invalidator InventoryInvalidator) {
	service.inventories = invalidator
}

// materialize derives current needs without writing during room reads or ticks.
func (service *Service) materialize(pet petrecord.Pet, now time.Time) petrecord.Pet {
	return petrecord.MaterializeStats(pet, now, service.config.StatDecayInterval, service.config.EnergyDecay, service.config.HappinessDecay)
}

// Start starts a fixed shared persistence pool.
func (service *Service) Start() {
	service.work.Start()
}

// Stop drains and stops shared persistence workers.
func (service *Service) Stop() {
	service.work.Stop()
}

// dispatch queues deferred work without a goroutine per pet.
func (service *Service) dispatch(job func()) bool {
	if service.work == nil {
		return false
	}
	return service.work.Dispatch(job)
}

// Publish emits one optional pet event.
func (service *Service) Publish(ctx context.Context, name bus.Name, payload any) {
	if service.events != nil {
		_ = service.events.Publish(ctx, bus.Event{Name: name, Payload: payload})
	}
}

// projectUnitStatus broadcasts one pet's current unit status.
func (service *Service) projectUnitStatus(ctx context.Context, active *roomlive.Room, petID int64) {
	service.projectEntityStatus(ctx, active, EntityKey(petID))
}

// projectEntityStatus broadcasts one entity's current unit status.
func (service *Service) projectEntityStatus(ctx context.Context, active *roomlive.Room, entityKey int64) {
	records := roomprojection.Statuses(active, entityKey)
	if len(records) == 0 {
		return
	}
	packet, err := outstatus.Encode(records)
	if err == nil {
		_ = broadcast.RoomPacket(ctx, service.connections, active, packet, 0)
	}
}
