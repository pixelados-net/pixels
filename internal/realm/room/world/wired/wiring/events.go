package wiring

import (
	"context"
	"strings"
	"time"

	botcore "github.com/niflaot/pixels/internal/realm/bot/core"
	furnituremoved "github.com/niflaot/pixels/internal/realm/furniture/events/moved"
	furniturepicked "github.com/niflaot/pixels/internal/realm/furniture/events/pickedup"
	furnitureplaced "github.com/niflaot/pixels/internal/realm/furniture/events/placed"
	furnitureused "github.com/niflaot/pixels/internal/realm/furniture/events/used"
	furnitureoff "github.com/niflaot/pixels/internal/realm/furniture/events/walkedoff"
	furnitureon "github.com/niflaot/pixels/internal/realm/furniture/events/walkedon"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	roommoved "github.com/niflaot/pixels/internal/realm/room/world/events/moved"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/game"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/record"
	wiredruntime "github.com/niflaot/pixels/internal/realm/room/world/wired/runtime"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/trigger"
	"github.com/niflaot/pixels/pkg/bus"
	"go.uber.org/zap"
)

// usedHandler emits state-changed after non-WIRED furniture interactions.
func usedHandler(rooms *roomlive.Registry, engine *wiredruntime.Engine) bus.Handler {
	return func(_ context.Context, event bus.Event) error {
		payload, ok := event.Payload.(furnitureused.Payload)
		if !ok {
			return nil
		}
		if active, found := rooms.Find(payload.RoomID); found {
			if item, exists := active.FurnitureItem(payload.ItemID); exists && strings.HasPrefix(item.Definition.InteractionType, "wf_") {
				return nil
			}
		}
		scheduleEvent(rooms, engine, unitEvent(rooms, trigger.StateChanged, payload.RoomID, payload.PlayerID, payload.ItemID))
		return nil
	}
}

// walkOnHandler emits completed walk-on context.
func walkOnHandler(rooms *roomlive.Registry, engine *wiredruntime.Engine, coordinator *game.Coordinator) bus.Handler {
	return func(ctx context.Context, event bus.Event) error {
		payload, ok := event.Payload.(furnitureon.Payload)
		if ok {
			if err := coordinator.Blob(ctx, payload); err != nil {
				return err
			}
			scheduleEvent(rooms, engine, unitEvent(rooms, trigger.WalkOn, payload.RoomID, payload.PlayerID, payload.ItemID))
		}
		return nil
	}
}

// walkOffHandler emits completed walk-off context.
func walkOffHandler(rooms *roomlive.Registry, engine *wiredruntime.Engine) bus.Handler {
	return func(_ context.Context, event bus.Event) error {
		payload, ok := event.Payload.(furnitureoff.Payload)
		if ok {
			scheduleEvent(rooms, engine, unitEvent(rooms, trigger.WalkOff, payload.RoomID, payload.PlayerID, payload.ItemID))
		}
		return nil
	}
}

// unitEvent creates player furniture event context.
func unitEvent(rooms *roomlive.Registry, kind trigger.Kind, roomID int64, playerID int64, itemID int64) trigger.Event {
	event := trigger.Event{Kind: kind, RoomID: roomID, ActorKind: trigger.ActorPlayer, ActorID: playerID, PlayerID: playerID, SourceItem: itemID}
	if active, found := rooms.Find(roomID); found {
		if actor, exists := active.UnitMotion(playerID); exists {
			event.ActorKind = actorKind(actor.Kind)
			event.ActorID = actor.EntityKey
			event.PlayerID = actor.PlayerID
		}
		if item, exists := active.FurnitureItem(itemID); exists {
			event.SourceSprite = int32(item.Definition.SpriteID)
		}
	}
	return event
}

// unitMovedHandler emits non-player arrival triggers from authoritative completed steps.
func unitMovedHandler(rooms *roomlive.Registry, bots *botcore.Service, engine *wiredruntime.Engine) bus.Handler {
	return func(_ context.Context, event bus.Event) error {
		payload, ok := event.Payload.(roommoved.Payload)
		if !ok || payload.Kind != worldunit.KindBot && payload.Kind != worldunit.KindPet {
			return nil
		}
		active, activeFound := rooms.Find(payload.RoomID)
		if !activeFound {
			return nil
		}
		username := ""
		if payload.Kind == worldunit.KindBot {
			if bots == nil {
				return nil
			}
			botID := payload.EntityKey
			if botID < 0 {
				botID = -botID
			}
			bot, found := bots.ResolveByID(payload.RoomID, botID)
			if !found {
				return nil
			}
			username = bot.Name
		}
		base := trigger.Event{RoomID: payload.RoomID, ActorKind: actorKind(payload.Kind), ActorID: payload.EntityKey, Username: username}
		var storage [16]int64
		for _, itemID := range active.FurnitureIDsAt(payload.Current.Point, storage[:0]) {
			arrival := base
			arrival.Kind, arrival.SourceItem = trigger.BotReachedFurniture, itemID
			if item, exists := active.FurnitureItem(itemID); exists {
				arrival.SourceSprite = int32(item.Definition.SpriteID)
			}
			scheduleEvent(rooms, engine, arrival)
		}
		for _, occupant := range active.Occupants() {
			unit, exists := active.UnitMotion(occupant.PlayerID)
			if exists && !nearAvatar(payload.Previous.Point, unit.Position.Point) && nearAvatar(payload.Current.Point, unit.Position.Point) {
				arrival := base
				arrival.Kind = trigger.BotReachedAvatar
				scheduleEvent(rooms, engine, arrival)
			}
		}
		return nil
	}
}

// actorKind maps world entity kinds into WIRED actor policies.
func actorKind(kind worldunit.Kind) trigger.ActorKind {
	switch kind {
	case worldunit.KindBot:
		return trigger.ActorBot
	case worldunit.KindPet:
		return trigger.ActorPet
	default:
		return trigger.ActorPlayer
	}
}

// nearAvatar reports classic bot reach distance below two tiles.
func nearAvatar(first grid.Point, second grid.Point) bool {
	dx, dy := int(first.X)-int(second.X), int(first.Y)-int(second.Y)
	if dx < 0 {
		dx = -dx
	}
	if dy < 0 {
		dy = -dy
	}
	return dx <= 1 && dy <= 1
}

// reloadMovedHandler recompiles after any furniture position change.
func reloadMovedHandler(engine *wiredruntime.Engine, log *zap.Logger) bus.Handler {
	return func(ctx context.Context, event bus.Event) error {
		payload, ok := event.Payload.(furnituremoved.Payload)
		if ok && engine.Loaded(payload.RoomID) {
			if err := engine.Reload(ctx, payload.RoomID, time.Now()); err != nil {
				logError(log, "reload moved room WIRED", payload.RoomID, err)
			}
		}
		return nil
	}
}

// reloadPlacedHandler recompiles after furniture placement.
func reloadPlacedHandler(engine *wiredruntime.Engine, log *zap.Logger) bus.Handler {
	return func(ctx context.Context, event bus.Event) error {
		payload, ok := event.Payload.(furnitureplaced.Payload)
		if ok && engine.Loaded(payload.RoomID) {
			if err := engine.Reload(ctx, payload.RoomID, time.Now()); err != nil {
				logError(log, "reload placed room WIRED", payload.RoomID, err)
			}
		}
		return nil
	}
}

// reloadPickedHandler recompiles after furniture pickup and cascade cleanup.
func reloadPickedHandler(store record.Store, engine *wiredruntime.Engine, log *zap.Logger) bus.Handler {
	return func(ctx context.Context, event bus.Event) error {
		payload, ok := event.Payload.(furniturepicked.Payload)
		if !ok {
			return nil
		}
		if err := store.CleanupItem(ctx, payload.ItemID); err != nil {
			logError(log, "clean picked room WIRED", payload.RoomID, err)
			return nil
		}
		if engine.Loaded(payload.RoomID) {
			if err := engine.Reload(ctx, payload.RoomID, time.Now()); err != nil {
				logError(log, "reload picked room WIRED", payload.RoomID, err)
			}
		}
		return nil
	}
}
