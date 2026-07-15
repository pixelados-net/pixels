// Package lifecycle owns durable bot inventory placement workflows.
package lifecycle

import (
	"context"

	"github.com/niflaot/pixels/internal/permission"
	permissionservice "github.com/niflaot/pixels/internal/permission/service"
	botcore "github.com/niflaot/pixels/internal/realm/bot/core"
	botpickedup "github.com/niflaot/pixels/internal/realm/bot/events/pickedup"
	botplaced "github.com/niflaot/pixels/internal/realm/bot/events/placed"
	botpolicy "github.com/niflaot/pixels/internal/realm/bot/policy"
	botrecord "github.com/niflaot/pixels/internal/realm/bot/record"
	furniturewalkedoff "github.com/niflaot/pixels/internal/realm/furniture/events/walkedoff"
	furniturewalkedon "github.com/niflaot/pixels/internal/realm/furniture/events/walkedon"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldpath "github.com/niflaot/pixels/internal/realm/room/world/path"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
)

// Service coordinates durable bot inventory and room placement.
type Service struct {
	// config stores bot limits.
	config botpolicy.Config
	// store persists bots.
	store botrecord.Store
	// rooms stores active rooms.
	rooms *roomlive.Registry
	// permissions resolves management bypasses.
	permissions permissionservice.Checker
	// runtime owns live bot state and protocol projection.
	runtime *botcore.Service
}

// New creates bot lifecycle behavior.
func New(config botpolicy.Config, store botrecord.Store, rooms *roomlive.Registry, permissions permissionservice.Checker, runtime *botcore.Service) *Service {
	return &Service{config: config.Normalize(), store: store, rooms: rooms, permissions: permissions, runtime: runtime}
}

// PlaceParams stores one player bot placement request.
type PlaceParams struct {
	// BotID identifies the owned inventory bot.
	BotID int64
	// ActorPlayerID identifies the placing player.
	ActorPlayerID int64
	// RoomID identifies the current room.
	RoomID int64
	// Point stores the requested tile.
	Point grid.Point
}

// PickupParams stores one player bot pickup request.
type PickupParams struct {
	// BotID identifies the placed bot.
	BotID int64
	// ActorPlayerID identifies the receiving player.
	ActorPlayerID int64
	// RoomID identifies the current room.
	RoomID int64
}

// Inventory returns one player's current bot inventory.
func (service *Service) Inventory(ctx context.Context, playerID int64) ([]botrecord.Bot, error) {
	return service.store.Inventory(ctx, playerID)
}

// Find returns one durable bot.
func (service *Service) Find(ctx context.Context, botID int64) (botrecord.Bot, bool, error) {
	return service.store.Find(ctx, botID)
}

// Place adds an owned bot to an active room atomically with persistence.
func (service *Service) Place(ctx context.Context, params PlaceParams) (botrecord.Bot, error) {
	bot, found, err := service.store.Find(ctx, params.BotID)
	if err != nil || !found || bot.OwnerPlayerID != params.ActorPlayerID || !bot.Inventory() {
		return botrecord.Bot{}, firstError(err, botrecord.ErrBotNotFound)
	}
	active, found := service.rooms.Find(params.RoomID)
	if !found {
		return botrecord.Bot{}, botrecord.ErrRoomNotFound
	}
	if err = service.runtime.EnsureRoom(ctx, active); err != nil {
		return botrecord.Bot{}, err
	}
	allowed, err := service.canPlace(ctx, active, params.ActorPlayerID)
	if err != nil || !allowed {
		return botrecord.Bot{}, firstError(err, botrecord.ErrNoRights)
	}
	if service.runtime.PlacedCount(params.RoomID) >= service.config.MaxPerRoom {
		unlimited, checkErr := service.has(ctx, params.ActorPlayerID, botpolicy.Unlimited)
		if checkErr != nil || !unlimited {
			return botrecord.Bot{}, firstError(checkErr, botrecord.ErrRoomLimit)
		}
	}
	position := worldpath.Position{Point: params.Point}
	unit, err := active.AddEntity(botcore.EntityKey(bot.ID), bot.OwnerPlayerID, worldunit.KindBot, position, worldunit.RotationSouth)
	if err != nil {
		return botrecord.Bot{}, botrecord.ErrTileNotFree
	}
	placed, found, err := service.store.Place(ctx, bot.ID, bot.OwnerPlayerID, params.RoomID, int(params.Point.X), int(params.Point.Y), unit.Position.Z.Units(), int16(unit.BodyRotation))
	if err != nil || !found {
		active.RemoveEntity(botcore.EntityKey(bot.ID))
		return botrecord.Bot{}, firstError(err, botrecord.ErrConflict)
	}
	service.runtime.AddPlaced(placed)
	service.runtime.ProjectSpawn(ctx, active, placed)
	service.runtime.SendInventoryRemove(ctx, params.ActorPlayerID, placed.ID)
	service.runtime.OnPlaced(ctx, placed)
	if item, itemFound := active.InteractionAt(params.Point); itemFound {
		service.runtime.Publish(ctx, furniturewalkedon.Name, furniturewalkedon.Payload{PlayerID: botcore.EntityKey(placed.ID), ItemID: item.ID, RoomID: params.RoomID})
	}
	service.runtime.Publish(ctx, botplaced.Name, botplaced.Payload{BotID: placed.ID, RoomID: params.RoomID, PlayerID: params.ActorPlayerID})
	return placed, nil
}

// Pickup removes a placed bot and gives it to the acting player's inventory.
func (service *Service) Pickup(ctx context.Context, params PickupParams) (botrecord.Bot, error) {
	bot, found, err := service.store.Find(ctx, params.BotID)
	if err != nil || !found || bot.RoomID == nil || *bot.RoomID != params.RoomID {
		return botrecord.Bot{}, firstError(err, botrecord.ErrBotNotFound)
	}
	allowed := bot.OwnerPlayerID == params.ActorPlayerID
	if !allowed {
		allowed, err = service.has(ctx, params.ActorPlayerID, botpolicy.AnyRoomOwner)
	}
	if err != nil || !allowed {
		return botrecord.Bot{}, firstError(err, botrecord.ErrNoRights)
	}
	count, err := service.store.CountInventory(ctx, params.ActorPlayerID)
	if err != nil {
		return botrecord.Bot{}, err
	}
	if count >= service.config.MaxInventory {
		unlimited, checkErr := service.has(ctx, params.ActorPlayerID, botpolicy.Unlimited)
		if checkErr != nil || !unlimited {
			return botrecord.Bot{}, firstError(checkErr, botrecord.ErrInventoryLimit)
		}
	}
	picked, found, err := service.store.Pickup(ctx, bot.ID, params.RoomID, params.ActorPlayerID)
	if err != nil || !found {
		return botrecord.Bot{}, firstError(err, botrecord.ErrConflict)
	}
	service.runtime.OnPickup(ctx, bot)
	service.runtime.RemovePlaced(params.RoomID, bot.ID)
	if active, activeFound := service.rooms.Find(params.RoomID); activeFound {
		if bot.X != nil && bot.Y != nil {
			if point, valid := grid.NewPoint(*bot.X, *bot.Y); valid {
				if item, itemFound := active.InteractionAt(point); itemFound {
					service.runtime.Publish(ctx, furniturewalkedoff.Name, furniturewalkedoff.Payload{PlayerID: botcore.EntityKey(bot.ID), ItemID: item.ID, RoomID: params.RoomID})
				}
			}
		}
		if unit, removed := active.RemoveEntity(botcore.EntityKey(bot.ID)); removed {
			service.runtime.ProjectRemove(ctx, active, unit.UnitID)
		}
	}
	service.runtime.SendInventoryAdd(ctx, params.ActorPlayerID, picked, true)
	service.runtime.Publish(ctx, botpickedup.Name, botpickedup.Payload{BotID: picked.ID, RoomID: params.RoomID, PlayerID: params.ActorPlayerID})
	return picked, nil
}

// Delete permanently removes an owned inventory bot.
func (service *Service) Delete(ctx context.Context, botID int64, ownerID int64) error {
	deleted, err := service.store.Delete(ctx, botID, ownerID)
	if err != nil {
		return err
	}
	if !deleted {
		return botrecord.ErrBotNotFound
	}
	return nil
}

// ForcePickup administratively returns a placed bot to its current owner.
func (service *Service) ForcePickup(ctx context.Context, botID int64) (botrecord.Bot, error) {
	bot, found, err := service.store.Find(ctx, botID)
	if err != nil || !found || bot.RoomID == nil {
		return botrecord.Bot{}, firstError(err, botrecord.ErrBotNotFound)
	}
	roomID := *bot.RoomID
	picked, found, err := service.store.ForcePickup(ctx, botID)
	if err != nil || !found {
		return botrecord.Bot{}, firstError(err, botrecord.ErrConflict)
	}
	service.runtime.OnPickup(ctx, bot)
	service.runtime.RemovePlaced(roomID, botID)
	if active, activeFound := service.rooms.Find(roomID); activeFound {
		if bot.X != nil && bot.Y != nil {
			if point, valid := grid.NewPoint(*bot.X, *bot.Y); valid {
				if item, itemFound := active.InteractionAt(point); itemFound {
					service.runtime.Publish(ctx, furniturewalkedoff.Name, furniturewalkedoff.Payload{PlayerID: botcore.EntityKey(bot.ID), ItemID: item.ID, RoomID: roomID})
				}
			}
		}
		if unit, removed := active.RemoveEntity(botcore.EntityKey(botID)); removed {
			service.runtime.ProjectRemove(ctx, active, unit.UnitID)
		}
	}
	service.runtime.SendInventoryAdd(ctx, picked.OwnerPlayerID, picked, true)
	return picked, nil
}

// canPlace resolves local room rights and explicit bot bypasses.
func (service *Service) canPlace(ctx context.Context, room *roomlive.Room, playerID int64) (bool, error) {
	if room != nil && room.CanManageFurniture(playerID) {
		return true, nil
	}
	if allowed, err := service.has(ctx, playerID, botpolicy.PlaceAnywhere); err != nil || allowed {
		return allowed, err
	}
	return service.has(ctx, playerID, botpolicy.AnyRoomOwner)
}

// has resolves one optional permission checker.
func (service *Service) has(ctx context.Context, playerID int64, node permission.Node) (bool, error) {
	if service.permissions == nil {
		return false, nil
	}
	return service.permissions.HasPermission(ctx, playerID, node)
}

// firstError chooses infrastructure errors before expected domain errors.
func firstError(err error, fallback error) error {
	if err != nil {
		return err
	}
	return fallback
}
