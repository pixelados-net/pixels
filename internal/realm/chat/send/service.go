// Package send validates and delivers room chat messages.
package send

import (
	"context"
	"time"

	"github.com/niflaot/pixels/internal/permission"
	permissionservice "github.com/niflaot/pixels/internal/permission/service"
	chatconfig "github.com/niflaot/pixels/internal/realm/chat/config"
	chatfilter "github.com/niflaot/pixels/internal/realm/chat/filter"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	roomwordfilter "github.com/niflaot/pixels/internal/realm/room/control/wordfilter"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	netconn "github.com/niflaot/pixels/networking/connection"
	"github.com/niflaot/pixels/pkg/bus"
	"github.com/niflaot/pixels/pkg/i18n"
	sdkplayer "github.com/niflaot/pixels/sdk/player"
)

// Counter increments expiring flood-control keys.
type Counter interface {
	// Increment increments a key and assigns its first-use expiration.
	Increment(context.Context, string, time.Duration) (int64, error)
}

// SpeechInterceptor examines filtered talk and shout messages before delivery.
type SpeechInterceptor interface {
	// Intercept returns whether WIRED consumed the normal room message.
	Intercept(context.Context, int64, int64, string, string) (bool, error)
}

// CommandExecutor handles command-prefixed room chat before normal validation.
type CommandExecutor interface {
	// Execute reports whether a message was consumed as a command.
	Execute(context.Context, sdkplayer.Player, string) (bool, error)
}

// EventDispatcher emits the cancellable pre-delivery plugin chat event.
type EventDispatcher interface {
	// DispatchChat returns the possibly replaced text and final cancellation state.
	DispatchChat(context.Context, sdkplayer.Player, int64, string) (string, bool)
}

// Nodes stores chat permission capabilities.
type Nodes struct {
	// FloodImmune exempts flood checks.
	FloodImmune permission.Node
	// LengthUnlimited exempts message length checks.
	LengthUnlimited permission.Node
	// FilterImmune exempts word filters.
	FilterImmune permission.Node
	// WhisperObserveAny allows passive whisper observation.
	WhisperObserveAny permission.Node
	// ModerationOwnMute allows room-local mute control.
	ModerationOwnMute permission.Node
	// ModerationAnyMute allows global room mute control.
	ModerationAnyMute permission.Node
}

// Service coordinates the allocation-sensitive room chat pipeline.
type Service struct {
	// config stores normalized chat limits.
	config chatconfig.Config
	// players stores authenticated live players.
	players *playerlive.Registry
	// bindings resolves source connections.
	bindings *binding.Registry
	// rooms stores active room worlds.
	rooms *roomlive.Registry
	// connections sends packets to active occupants.
	connections *netconn.Registry
	// permissions resolves chat bypasses.
	permissions permissionservice.Checker
	// counter stores cross-room flood counters.
	counter Counter
	// globalFilter applies the hotel dictionary.
	globalFilter *chatfilter.Service
	// roomFilter applies room-specific dictionaries.
	roomFilter roomwordfilter.Manager
	// events publishes delivered message events.
	events bus.Publisher
	// translations resolves end-user feedback.
	translations i18n.Translator
	// speechInterceptor bridges filtered speech into WIRED triggers.
	speechInterceptor SpeechInterceptor
	// commandExecutor handles plugin command chat before normal delivery.
	commandExecutor CommandExecutor
	// eventDispatcher emits cancellable plugin chat events.
	eventDispatcher EventDispatcher
	// nodes stores capability identifiers.
	nodes Nodes
	// now returns current time for deterministic checks.
	now func() time.Time
}

// SetSpeechInterceptor installs the WIRED speech bridge.
func (service *Service) SetSpeechInterceptor(interceptor SpeechInterceptor) {
	service.speechInterceptor = interceptor
}

// SetPluginRuntime installs plugin command and event dispatch seams.
func (service *Service) SetPluginRuntime(commands CommandExecutor, events EventDispatcher) {
	service.commandExecutor = commands
	service.eventDispatcher = events
}

// New creates a room chat service.
func New(config chatconfig.Config, players *playerlive.Registry, bindings *binding.Registry, rooms *roomlive.Registry, connections *netconn.Registry, permissions permissionservice.Checker, counter Counter, globalFilter *chatfilter.Service, roomFilter roomwordfilter.Manager, events bus.Publisher, translations i18n.Translator, nodes Nodes) *Service {
	return &Service{
		config: config.Normalize(), players: players, bindings: bindings, rooms: rooms,
		connections: connections, permissions: permissions, counter: counter,
		globalFilter: globalFilter, roomFilter: roomFilter, events: events,
		translations: translations, nodes: nodes, now: time.Now,
	}
}
