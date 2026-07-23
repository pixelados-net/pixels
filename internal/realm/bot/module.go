package bot

import (
	"context"

	botadmin "github.com/niflaot/pixels/internal/realm/bot/admin"
	"github.com/niflaot/pixels/internal/realm/bot/behavior"
	botcommands "github.com/niflaot/pixels/internal/realm/bot/commands"
	botcore "github.com/niflaot/pixels/internal/realm/bot/core"
	botdb "github.com/niflaot/pixels/internal/realm/bot/database"
	bothandlers "github.com/niflaot/pixels/internal/realm/bot/handlers"
	botlifecycle "github.com/niflaot/pixels/internal/realm/bot/lifecycle"
	botpolicy "github.com/niflaot/pixels/internal/realm/bot/policy"
	botrecord "github.com/niflaot/pixels/internal/realm/bot/record"
	botsettings "github.com/niflaot/pixels/internal/realm/bot/settings"
	chatshouted "github.com/niflaot/pixels/internal/realm/chat/events/shouted"
	chattalked "github.com/niflaot/pixels/internal/realm/chat/events/talked"
	realmconn "github.com/niflaot/pixels/internal/realm/connection"
	playerdisconnected "github.com/niflaot/pixels/internal/realm/player/events/disconnected"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	roomentered "github.com/niflaot/pixels/internal/realm/room/access/events/entered"
	roomleft "github.com/niflaot/pixels/internal/realm/room/access/events/left"
	roombundle "github.com/niflaot/pixels/internal/realm/room/record/bundle"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	wiredbridge "github.com/niflaot/pixels/internal/realm/room/world/wired/bridge"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	"github.com/niflaot/pixels/pkg/bus"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Module provides bot persistence, behaviors, protocol, and room runtime wiring.
var Module = fx.Module(
	"realm-bot",
	fx.Provide(botpolicy.LoadConfig, botdb.New, NewStore, NewRoomBundleCloner, behavior.NewRegistry, botcore.New, botlifecycle.New, botsettings.New, botadmin.New, NewInventoryHandler, NewPlacementHandler, NewSettingsHandler),
	fx.Invoke(behavior.RegisterBuiltins),
	fx.Invoke(RegisterHandlers),
	fx.Invoke(RegisterRuntime),
	fx.Invoke(RegisterWiredSpeech),
)

// RegisterWiredSpeech installs filtered bot-speech interception.
func RegisterWiredSpeech(service *botcore.Service, bridge *wiredbridge.SpeechBridge) {
	service.SetSpeechInterceptor(wiredbridge.NewBotSpeechInterceptor(bridge))
}

// NewStore exposes bot persistence through its domain contract.
func NewStore(repository *botdb.Repository) botrecord.Store { return repository }

// NewRoomBundleCloner exposes set-based bot cloning to room bundles.
func NewRoomBundleCloner(repository *botdb.Repository) roombundle.BotCloner { return repository }

// HandlerDeps contains bot packet handler dependencies.
type HandlerDeps struct {
	fx.In
	// Handlers stores the realm inbound packet registry.
	Handlers *realmconn.Handlers
	// Inventory handles bot inventory reads.
	Inventory botcommands.InventoryHandler
	// Placement handles bot placement and pickup.
	Placement botcommands.PlacementHandler
	// Settings handles bot configuration.
	Settings botcommands.SettingsHandler
	// Log records command dispatch.
	Log *zap.Logger
}

// CommandDeps contains shared grouped bot command dependencies.
type CommandDeps struct {
	fx.In
	// Lifecycle coordinates inventory placement behavior.
	Lifecycle *botlifecycle.Service
	// Settings coordinates bot skill configuration.
	Settings *botsettings.Service
	// Players stores live player state.
	Players *playerlive.Registry
	// Bindings stores authenticated connection bindings.
	Bindings *binding.Registry
}

// RegisterHandlers installs every bot packet adapter.
func RegisterHandlers(deps HandlerDeps) {
	if deps.Handlers == nil || deps.Handlers.Inbound == nil {
		return
	}
	bothandlers.RegisterInventory(deps.Handlers.Inbound, bothandlers.NewInventory(deps.Inventory, deps.Log))
	bothandlers.RegisterPlacement(deps.Handlers.Inbound, bothandlers.NewPlace(deps.Placement, deps.Log), bothandlers.NewPickup(deps.Placement, deps.Log))
	bothandlers.RegisterSettings(deps.Handlers.Inbound, bothandlers.NewConfiguration(deps.Settings, deps.Log), bothandlers.NewSkillSave(deps.Settings, deps.Log))
}

// NewInventoryHandler creates the grouped bot inventory command handler.
func NewInventoryHandler(params CommandDeps) botcommands.InventoryHandler {
	return botcommands.InventoryHandler{Service: params.Lifecycle, Players: params.Players, Bindings: params.Bindings}
}

// NewPlacementHandler creates the grouped bot placement command handler.
func NewPlacementHandler(params CommandDeps) botcommands.PlacementHandler {
	return botcommands.PlacementHandler{Service: params.Lifecycle, Players: params.Players, Bindings: params.Bindings}
}

// NewSettingsHandler creates the grouped bot settings command handler.
func NewSettingsHandler(params CommandDeps) botcommands.SettingsHandler {
	return botcommands.SettingsHandler{Service: params.Settings, Players: params.Players, Bindings: params.Bindings}
}

// RegisterRuntime attaches bots to the shared room cycle and realm events.
func RegisterRuntime(lifecycle fx.Lifecycle, subscriber bus.Subscriber, rooms *roomlive.Registry, service *botcore.Service) error {
	rooms.AddCyclePublisher(service.Cycle)
	rooms.AddClosePublisher(service.UnloadRoom)
	subscriptions, err := subscribeRuntime(subscriber, service)
	if err != nil {
		return err
	}
	lifecycle.Append(fx.Hook{
		OnStart: func(context.Context) error { service.Start(); return nil },
		OnStop: func(context.Context) error {
			for _, subscription := range subscriptions {
				subscription.Unsubscribe()
			}
			service.Stop()
			return nil
		},
	})
	return nil
}

// subscribeRuntime registers bot room lifecycle and chat handlers.
func subscribeRuntime(subscriber bus.Subscriber, service *botcore.Service) ([]*bus.Subscription, error) {
	registrations := []struct {
		name   bus.Name
		handle bus.Handler
	}{
		{name: roomentered.Name, handle: func(ctx context.Context, event bus.Event) error {
			payload, ok := event.Payload.(roomentered.Payload)
			if !ok {
				return nil
			}
			return service.HandleUserEnter(ctx, payload.RoomID, payload.PlayerID)
		}},
		{name: roomleft.Name, handle: func(_ context.Context, event bus.Event) error {
			payload, ok := event.Payload.(roomleft.Payload)
			if ok {
				service.StopFollowing(payload.RoomID, payload.PlayerID)
			}
			return nil
		}},
		{name: playerdisconnected.Name, handle: func(_ context.Context, event bus.Event) error {
			payload, ok := event.Payload.(playerdisconnected.Payload)
			if !ok {
				return nil
			}
			service.StopFollowingEverywhere(payload.PlayerID)
			return nil
		}},
		{name: chattalked.Name, handle: userSayHandler(service)},
		{name: chatshouted.Name, handle: userShoutHandler(service)},
	}
	result := make([]*bus.Subscription, 0, len(registrations))
	for _, registration := range registrations {
		subscription, err := subscriber.Subscribe(registration.name, bus.PriorityNormal, registration.handle)
		if err != nil {
			for _, current := range result {
				current.Unsubscribe()
			}
			return nil, err
		}
		result = append(result, subscription)
	}
	return result, nil
}

// userSayHandler adapts ordinary room chat into asynchronous behavior input.
func userSayHandler(service *botcore.Service) bus.Handler {
	return func(ctx context.Context, event bus.Event) error {
		payload, ok := event.Payload.(chattalked.Payload)
		if ok {
			service.HandleUserSay(ctx, payload.RoomID, payload.PlayerID, payload.Message)
		}
		return nil
	}
}

// userShoutHandler adapts room shouts into asynchronous behavior input.
func userShoutHandler(service *botcore.Service) bus.Handler {
	return func(ctx context.Context, event bus.Event) error {
		payload, ok := event.Payload.(chatshouted.Payload)
		if ok {
			service.HandleUserSay(ctx, payload.RoomID, payload.PlayerID, payload.Message)
		}
		return nil
	}
}
