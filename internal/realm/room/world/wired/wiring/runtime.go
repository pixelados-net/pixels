package wiring

import (
	"context"
	"time"

	botcore "github.com/niflaot/pixels/internal/realm/bot/core"
	furnituremoved "github.com/niflaot/pixels/internal/realm/furniture/events/moved"
	furniturepicked "github.com/niflaot/pixels/internal/realm/furniture/events/pickedup"
	furnitureplaced "github.com/niflaot/pixels/internal/realm/furniture/events/placed"
	furnitureused "github.com/niflaot/pixels/internal/realm/furniture/events/used"
	furnitureoff "github.com/niflaot/pixels/internal/realm/furniture/events/walkedoff"
	furnitureon "github.com/niflaot/pixels/internal/realm/furniture/events/walkedon"
	socialgroup "github.com/niflaot/pixels/internal/realm/group"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	roomentered "github.com/niflaot/pixels/internal/realm/room/access/events/entered"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	roommoved "github.com/niflaot/pixels/internal/realm/room/world/events/moved"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/game"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/record"
	wiredruntime "github.com/niflaot/pixels/internal/realm/room/world/wired/runtime"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/trigger"
	"github.com/niflaot/pixels/pkg/bus"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// RegisterRuntime connects WIRED to room lifecycle and realm events.
func RegisterRuntime(lifecycle fx.Lifecycle, subscriber bus.Subscriber, rooms *roomlive.Registry, players *playerlive.Registry, bots *botcore.Service, engine *wiredruntime.Engine, games *game.Service, coordinator *game.Coordinator, groups *socialgroup.Service, store record.Store, log *zap.Logger) error {
	rooms.AddCyclePublisher(func(ctx context.Context, active *roomlive.Room, now time.Time) error {
		return engine.Cycle(ctx, active.ID(), now)
	})
	rooms.AddClosePublisher(func(roomID int64) {
		engine.Close(roomID)
		games.Close(roomID)
		groups.CloseRoom(roomID)
	})
	subscriptions, err := subscribeRuntime(subscriber, rooms, players, bots, engine, coordinator, groups, store, log)
	if err != nil {
		return err
	}
	lifecycle.Append(fx.Hook{OnStop: func(context.Context) error {
		for _, subscription := range subscriptions {
			subscription.Unsubscribe()
		}
		return nil
	}})
	return nil
}

// subscribeRuntime registers cold reloads and hot event adapters.
func subscribeRuntime(subscriber bus.Subscriber, rooms *roomlive.Registry, players *playerlive.Registry, bots *botcore.Service, engine *wiredruntime.Engine, coordinator *game.Coordinator, groups *socialgroup.Service, store record.Store, log *zap.Logger) ([]*bus.Subscription, error) {
	registrations := []struct {
		name   bus.Name
		handle bus.Handler
	}{
		{name: roomentered.Name, handle: enteredHandler(rooms, players, engine, groups, log)},
		{name: furnitureused.Name, handle: usedHandler(rooms, engine)},
		{name: furnitureon.Name, handle: walkOnHandler(rooms, engine, coordinator)},
		{name: furnitureoff.Name, handle: walkOffHandler(rooms, engine)},
		{name: furnituremoved.Name, handle: reloadMovedHandler(engine, log)},
		{name: furnitureplaced.Name, handle: reloadPlacedHandler(engine, log)},
		{name: furniturepicked.Name, handle: reloadPickedHandler(store, engine, log)},
		{name: roommoved.Name, handle: unitMovedHandler(rooms, bots, engine)},
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

// enteredHandler loads a generation once and emits the completed entry event.
func enteredHandler(rooms *roomlive.Registry, players *playerlive.Registry, engine *wiredruntime.Engine, groups *socialgroup.Service, log *zap.Logger) bus.Handler {
	return func(ctx context.Context, event bus.Event) error {
		payload, ok := event.Payload.(roomentered.Payload)
		if !ok {
			return nil
		}
		if !engine.Loaded(payload.RoomID) {
			if err := groups.PrepareRoom(ctx, payload.RoomID); err != nil {
				logError(log, "load room social group", payload.RoomID, err)
				return nil
			}
			if err := engine.Reload(ctx, payload.RoomID, time.Now()); err != nil {
				logError(log, "load room WIRED", payload.RoomID, err)
				return nil
			}
		}
		if err := groups.PreparePlayer(ctx, payload.PlayerID); err != nil {
			logError(log, "load player social groups", payload.RoomID, err)
			return nil
		}
		username := ""
		if player, found := players.Find(payload.PlayerID); found {
			username = player.Username()
		}
		scheduleEvent(rooms, engine, trigger.Event{Kind: trigger.EnterRoom, RoomID: payload.RoomID, ActorKind: trigger.ActorPlayer, ActorID: payload.PlayerID, PlayerID: payload.PlayerID, Username: username})
		return nil
	}
}

// scheduleEvent serializes event execution through the existing room task queue.
func scheduleEvent(rooms *roomlive.Registry, engine *wiredruntime.Engine, event trigger.Event) {
	active, found := rooms.Find(event.RoomID)
	if !found {
		return
	}
	active.Schedule(0, func(now time.Time) { _, _ = engine.Process(context.Background(), event, now) })
}

// logError records cold-path compilation failures without disconnecting a player.
func logError(log *zap.Logger, message string, roomID int64, err error) {
	if log != nil {
		log.Error(message, zap.Int64("room_id", roomID), zap.Error(err))
	}
}
