package core

import (
	"math/rand/v2"
	"sync"
	"sync/atomic"
	"time"

	permissionservice "github.com/niflaot/pixels/internal/permission/service"
	botbehavior "github.com/niflaot/pixels/internal/realm/bot/behavior"
	botpolicy "github.com/niflaot/pixels/internal/realm/bot/policy"
	botrecord "github.com/niflaot/pixels/internal/realm/bot/record"
	chatfilter "github.com/niflaot/pixels/internal/realm/chat/filter"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	playerservice "github.com/niflaot/pixels/internal/realm/player/service"
	roomwordfilter "github.com/niflaot/pixels/internal/realm/room/control/wordfilter"
	roomservice "github.com/niflaot/pixels/internal/realm/room/record/service"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	netconn "github.com/niflaot/pixels/networking/connection"
	"github.com/niflaot/pixels/pkg/bus"
	"github.com/niflaot/pixels/pkg/i18n"
	sdkbot "github.com/niflaot/pixels/sdk/bot"
	"go.uber.org/zap"
)

// Source supplies deterministic random values in tests.
type Source interface {
	// Uint64 returns one pseudo-random value.
	Uint64() uint64
}

// globalSource uses Go's concurrency-safe top-level random generator.
type globalSource struct{}

// Uint64 returns one pseudo-random value.
func (globalSource) Uint64() uint64 { return rand.Uint64() }

// activeBot stores one mutable placed bot schedule.
type activeBot struct {
	// id stores the immutable durable identifier used by the hot path.
	id int64
	// mutex protects mutable runtime fields.
	mutex sync.Mutex
	// behavior stores the behavior instance created for this placement.
	behavior sdkbot.Behavior
	// record stores the latest durable snapshot.
	record botrecord.Bot
	// chatIndex stores the next sequential chat line.
	chatIndex int
	// nextChat stores the next allowed automatic chat time.
	nextChat time.Time
	// nextWalk stores the next allowed random walk time.
	nextWalk time.Time
	// followingPlayerID stores the active follow target.
	followingPlayerID int64
	// visitorShown stores whether visitor detail was already shown this placement.
	visitorShown bool
	// visitorPrompted stores whether the owner received the visit summary this placement.
	visitorPrompted bool
	// lastPositionFlush stores deferred persistence cadence.
	lastPositionFlush time.Time
}

// roomState stores loaded bots and stable template variables for one room.
type roomState struct {
	// bots stores active bots by durable id.
	bots map[int64]*activeBot
	// snapshot stores an immutable zero-allocation cycle view.
	snapshot atomic.Pointer[botSnapshot]
	// name stores the room name snapshot.
	name string
	// ownerName stores the room owner snapshot.
	ownerName string
}

// botSnapshot stores one immutable active-bot pointer generation.
type botSnapshot struct {
	// bots stores stable active bot pointers.
	bots []*activeBot
}

// rebuildSnapshot publishes an immutable active bot pointer generation.
func (state *roomState) rebuildSnapshot() {
	bots := make([]*activeBot, 0, len(state.bots))
	for _, bot := range state.bots {
		bots = append(bots, bot)
	}
	state.snapshot.Store(&botSnapshot{bots: bots})
}

// EntityKey returns the collision-free negative world key for one bot.
func EntityKey(botID int64) int64 {
	if botID < 0 {
		return botID
	}
	return -botID
}

// Service coordinates durable and active bot state.
type Service struct {
	// config stores normalized behavior limits.
	config botpolicy.Config
	// store persists bot records.
	store botrecord.Store
	// rooms stores active room worlds.
	rooms *roomlive.Registry
	// roomRecords resolves room template variables.
	roomRecords roomservice.Finder
	// players resolves live identity and look state.
	players *playerlive.Registry
	// playerRecords resolves durable last-online state for visitor bots.
	playerRecords playerservice.AdminManager
	// permissions resolves bypass capabilities.
	permissions permissionservice.Checker
	// behaviors resolves registered behavior factories.
	behaviors *botbehavior.Registry
	// speechInterceptor provides the explicit Wired speech extension boundary.
	speechInterceptor sdkbot.SpeechInterceptor
	// connections sends Nitro projections.
	connections *netconn.Registry
	// globalFilter applies hotel chat moderation.
	globalFilter *chatfilter.Service
	// roomFilter applies room-specific moderation.
	roomFilter roomwordfilter.Manager
	// events publishes bot lifecycle events.
	events bus.Publisher
	// translations resolves hotel-facing bot messages.
	translations i18n.Translator
	// source supplies random decisions.
	source Source
	// log records deferred behavior failures.
	log *zap.Logger
	// mutex protects room state generations.
	mutex sync.RWMutex
	// active stores loaded bot state by room id.
	active map[int64]*roomState
	// serveItems stores cached bartender mappings.
	serveItems []botrecord.ServeItem
	// serveLoaded reports whether serving mappings were loaded.
	serveLoaded bool
	// jobs runs slow hooks and persistence away from room owner loops.
	jobs chan func()
	// stopped closes worker loops.
	stopped chan struct{}
	// startOnce prevents duplicate worker pools.
	startOnce sync.Once
	// stopOnce prevents duplicate shutdown signals.
	stopOnce sync.Once
	// workers waits for shared workers to stop.
	workers sync.WaitGroup
}

// New creates bot core behavior.
func New(config botpolicy.Config, store botrecord.Store, rooms *roomlive.Registry, roomRecords roomservice.Manager, players *playerlive.Registry, playerRecords playerservice.AdminManager, permissions permissionservice.Checker, behaviors *botbehavior.Registry, connections *netconn.Registry, globalFilter *chatfilter.Service, roomFilter roomwordfilter.Manager, events bus.Publisher, translations i18n.Translator, log *zap.Logger) *Service {
	return &Service{config: config.Normalize(), store: store, rooms: rooms, roomRecords: roomRecords, players: players, playerRecords: playerRecords, permissions: permissions, behaviors: behaviors, speechInterceptor: sdkbot.NoopSpeechInterceptor{}, connections: connections, globalFilter: globalFilter, roomFilter: roomFilter, events: events, translations: translations, source: globalSource{}, log: log, active: make(map[int64]*roomState), jobs: make(chan func(), 1024), stopped: make(chan struct{})}
}

// SetSpeechInterceptor installs the room speech bridge used by WIRED bot triggers.
func (service *Service) SetSpeechInterceptor(interceptor sdkbot.SpeechInterceptor) {
	if interceptor == nil {
		interceptor = sdkbot.NoopSpeechInterceptor{}
	}
	service.speechInterceptor = interceptor
}

// Start starts the fixed shared behavior worker pool.
func (service *Service) Start() {
	service.startOnce.Do(func() {
		service.workers.Add(4)
		for range 4 {
			go service.worker()
		}
	})
}

// Stop stops shared behavior workers.
func (service *Service) Stop() {
	service.stopOnce.Do(func() { close(service.stopped) })
	service.workers.Wait()
}

// worker executes bounded asynchronous behavior work.
func (service *Service) worker() {
	defer service.workers.Done()
	for {
		select {
		case job := <-service.jobs:
			if job != nil {
				job()
			}
		case <-service.stopped:
			service.drainJobs()
			return
		}
	}
}

// drainJobs completes work accepted before shutdown without blocking for new work.
func (service *Service) drainJobs() {
	for {
		select {
		case job := <-service.jobs:
			if job != nil {
				job()
			}
		default:
			return
		}
	}
}

// dispatch queues work without creating a goroutine per bot or event.
func (service *Service) dispatch(job func()) {
	select {
	case <-service.stopped:
		return
	case service.jobs <- job:
	default:
		if service.log != nil {
			service.log.Warn("bot behavior queue full")
		}
	}
}
