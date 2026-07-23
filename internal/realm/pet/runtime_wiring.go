package pet

import (
	"context"

	chatshouted "github.com/niflaot/pixels/internal/realm/chat/events/shouted"
	chattalked "github.com/niflaot/pixels/internal/realm/chat/events/talked"
	furniturepickedup "github.com/niflaot/pixels/internal/realm/furniture/events/pickedup"
	petbehavior "github.com/niflaot/pixels/internal/realm/pet/behavior"
	petbreeding "github.com/niflaot/pixels/internal/realm/pet/breeding"
	petequipment "github.com/niflaot/pixels/internal/realm/pet/equipment"
	petinventory "github.com/niflaot/pixels/internal/realm/pet/inventory"
	petreference "github.com/niflaot/pixels/internal/realm/pet/reference"
	petruntime "github.com/niflaot/pixels/internal/realm/pet/runtime"
	playerdisconnected "github.com/niflaot/pixels/internal/realm/player/events/disconnected"
	roomentered "github.com/niflaot/pixels/internal/realm/room/access/events/entered"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/pkg/bus"
	"go.uber.org/fx"
)

// RegisterRuntime attaches pets to the shared room cycle and realm events.
func RegisterRuntime(lifecycle fx.Lifecycle, subscriber bus.Subscriber, rooms *roomlive.Registry, references *petreference.Service, registry *petbehavior.Registry, runtime *petruntime.Service, behavior *petbehavior.Service, equipment *petequipment.Service, inventories *petinventory.Service, breeding *petbreeding.Service) error {
	if err := registry.Validate(); err != nil {
		return err
	}
	runtime.SetNeedConsumer(equipment)
	runtime.SetInventoryInvalidator(inventories)
	rooms.AddCyclePublisher(runtime.Cycle)
	rooms.AddClosePublisher(breeding.CloseRoom)
	rooms.AddClosePublisher(runtime.UnloadRoom)
	subscriptions, err := subscribeRuntime(subscriber, runtime, behavior, breeding)
	if err != nil {
		return err
	}
	lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if err := references.Refresh(ctx); err != nil {
				return err
			}
			runtime.Start()
			return nil
		},
		OnStop: func(context.Context) error {
			for _, subscription := range subscriptions {
				subscription.Unsubscribe()
			}
			runtime.Stop()
			return nil
		},
	})
	return nil
}

// subscribeRuntime registers pet room lifecycle and chat observers.
func subscribeRuntime(subscriber bus.Subscriber, runtime *petruntime.Service, behavior *petbehavior.Service, breeding *petbreeding.Service) ([]*bus.Subscription, error) {
	registrations := []struct {
		name    bus.Name
		handler bus.Handler
	}{
		{name: roomentered.Name, handler: petRoomEntered(runtime)},
		{name: playerdisconnected.Name, handler: petPlayerDisconnected(runtime)},
		{name: chattalked.Name, handler: petTalked(behavior)},
		{name: chatshouted.Name, handler: petShouted(behavior)},
		{name: furniturepickedup.Name, handler: petFurniturePickedUp(breeding)},
	}
	result := make([]*bus.Subscription, 0, len(registrations))
	for _, registration := range registrations {
		subscription, err := subscriber.Subscribe(registration.name, bus.PriorityNormal, registration.handler)
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

// petFurniturePickedUp cancels a nest session when its furniture leaves the room.
func petFurniturePickedUp(breeding *petbreeding.Service) bus.Handler {
	return func(ctx context.Context, event bus.Event) error {
		payload, ok := event.Payload.(furniturepickedup.Payload)
		if ok {
			_ = breeding.Cancel(ctx, payload.ItemID, payload.RoomID)
		}
		return nil
	}
}

// petRoomEntered synchronizes pets to one late room entrant.
func petRoomEntered(runtime *petruntime.Service) bus.Handler {
	return func(ctx context.Context, event bus.Event) error {
		payload, ok := event.Payload.(roomentered.Payload)
		if !ok {
			return nil
		}
		return runtime.SyncPlayer(ctx, payload.RoomID, payload.PlayerID)
	}
}

// petPlayerDisconnected dismounts one disconnected player.
func petPlayerDisconnected(runtime *petruntime.Service) bus.Handler {
	return func(ctx context.Context, event bus.Event) error {
		payload, ok := event.Payload.(playerdisconnected.Payload)
		if ok {
			runtime.DismountPlayer(ctx, payload.PlayerID)
		}
		return nil
	}
}

// petTalked forwards ordinary speech to pet command parsing.
func petTalked(behavior *petbehavior.Service) bus.Handler {
	return func(ctx context.Context, event bus.Event) error {
		payload, ok := event.Payload.(chattalked.Payload)
		if ok {
			_ = behavior.HandleSpeech(ctx, payload.RoomID, payload.PlayerID, payload.Message)
		}
		return nil
	}
}

// petShouted forwards shouted speech to pet command parsing.
func petShouted(behavior *petbehavior.Service) bus.Handler {
	return func(ctx context.Context, event bus.Event) error {
		payload, ok := event.Payload.(chatshouted.Payload)
		if ok {
			_ = behavior.HandleSpeech(ctx, payload.RoomID, payload.PlayerID, payload.Message)
		}
		return nil
	}
}
