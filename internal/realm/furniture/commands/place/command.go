// Package place places an inventory furniture item into the player's current room.
package place

import (
	"context"
	"errors"

	"github.com/niflaot/pixels/internal/command"
	permissionservice "github.com/niflaot/pixels/internal/permission/service"
	furnitureaccess "github.com/niflaot/pixels/internal/realm/furniture/access"
	furnituresession "github.com/niflaot/pixels/internal/realm/furniture/commands/session"
	placedevent "github.com/niflaot/pixels/internal/realm/furniture/events/placed"
	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	playerservice "github.com/niflaot/pixels/internal/realm/player/service"
	"github.com/niflaot/pixels/internal/realm/room/runtime/broadcast"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	worldfurniture "github.com/niflaot/pixels/internal/realm/room/world/furniture"
	roomfurniture "github.com/niflaot/pixels/internal/realm/room/world/items"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	netconn "github.com/niflaot/pixels/networking/connection"
	outinvremove "github.com/niflaot/pixels/networking/outbound/inventory/furniture/remove"
	outadd "github.com/niflaot/pixels/networking/outbound/room/furniture/add"
	"github.com/niflaot/pixels/pkg/bus"
	"github.com/niflaot/pixels/pkg/i18n"
	"go.uber.org/zap"
)

const (
	// Name identifies the furniture place command.
	Name command.Name = "furniture.place"
)

// ErrPlayerNotInRoom reports a placement attempt without active room presence.
var ErrPlayerNotInRoom = errors.New("player not in room")

// Command places an inventory furniture item.
type Command struct {
	// Handler stores the source connection handler.
	Handler netconn.Context

	// ItemID identifies the inventory furniture item to place.
	ItemID int64

	// X stores the destination floor tile x coordinate.
	X int

	// Y stores the destination floor tile y coordinate.
	Y int

	// Rotation stores the destination floor instance rotation.
	Rotation int

	// WallPosition stores Nitro wall coordinates for a wall item.
	WallPosition string
}

// Handler handles furniture place commands.
type Handler struct {
	// Players stores live player state.
	Players *playerlive.Registry

	// Bindings stores player connection bindings.
	Bindings *binding.Registry

	// Furniture manages placed and inventory furniture records.
	Furniture furnitureservice.Manager

	// PlayerDirectory resolves durable player identities for gift sender tags.
	PlayerDirectory playerservice.Finder

	// Runtime stores active rooms.
	Runtime *roomlive.Registry

	// Permissions resolves global furniture management authority.
	Permissions permissionservice.Checker

	// Connections stores active network connections.
	Connections *netconn.Registry

	// Events publishes furniture lifecycle events.
	Events bus.Publisher

	// Groups resolves warmed linked social-group furniture identity.
	Groups furnituremodel.GroupPolicy

	// Translations resolves end-user messages.
	Translations i18n.Translator

	// Log records rejected placement attempts.
	Log *zap.Logger

	// RollerNoRules disables roller-on-furniture placement restrictions.
	RollerNoRules bool
}

// CommandName returns the stable command name.
func (Command) CommandName() command.Name {
	return Name
}

// Handle handles a furniture place command.
func (handler Handler) Handle(ctx context.Context, envelope command.Envelope[Command]) error {
	player, err := furnituresession.Player(envelope.Command.Handler, handler.Bindings, handler.Players)
	if err != nil {
		return err
	}
	roomID, found := player.CurrentRoom()
	if !found {
		return nil
	}
	active, found := handler.Runtime.Find(roomID)
	if !found {
		return nil
	}
	allowed, err := furnitureaccess.CanManage(ctx, handler.Permissions, active, player.ID())
	if err != nil {
		return err
	}
	if !allowed {
		return handler.handleSoftError(ctx, envelope.Command, roomlive.ErrNoFurnitureRights)
	}

	item, found, err := handler.Furniture.FindItemByID(ctx, envelope.Command.ItemID)
	if err != nil {
		return err
	}
	if !found {
		return nil
	}
	definition, found, err := handler.Furniture.FindDefinitionByID(ctx, item.DefinitionID)
	if err != nil {
		return err
	}
	if !found {
		return handler.handleSoftError(ctx, envelope.Command, roomfurniture.ErrDefinitionNotFound)
	}
	if definition.Kind == furnituremodel.KindWall {
		return handler.placeWall(ctx, envelope.Command, player, active, roomID, item, definition)
	}

	rotation := furnituremodel.Rotation(envelope.Command.Rotation)
	worldItem, definition, err := roomfurniture.ResolveWorldItem(ctx, active, handler.Furniture, item, envelope.Command.X, envelope.Command.Y, rotation)
	if err != nil {
		return handler.handleSoftError(ctx, envelope.Command, err)
	}
	if err = roomfurniture.ValidateRollerPlacement(active, worldItem, handler.RollerNoRules); err != nil {
		return handler.handleSoftError(ctx, envelope.Command, err)
	}

	placed, err := handler.Furniture.Place(ctx, furnitureservice.PlaceParams{
		ItemID:        item.ID,
		ActorPlayerID: player.ID(),
		RoomID:        roomID,
		Placement: furnituremodel.Placement{
			X: envelope.Command.X, Y: envelope.Command.Y,
			Z: worldItem.Z.Units(), Rotation: rotation,
		},
	})
	if err != nil {
		return handler.handleSoftError(ctx, envelope.Command, err)
	}

	return project(ctx, func(projectionCtx context.Context) error {
		if _, reloadErr := active.ReloadFurniture(placed.ID, &worldItem); reloadErr != nil {
			return reloadErr
		}
		if sendErr := handler.sendInventoryRemove(projectionCtx, envelope.Command.Handler, placed.ID); sendErr != nil {
			return sendErr
		}
		if broadcastErr := handler.broadcastAdd(projectionCtx, active, placed, definition, player.Username()); broadcastErr != nil {
			return broadcastErr
		}
		if heightErr := handler.broadcastHeightMapUpdate(projectionCtx, active, worldItem); heightErr != nil {
			return heightErr
		}
		if handler.Log != nil {
			handler.Log.Debug("furniture placed",
				zap.Int64("player_id", player.ID()), zap.Int64("item_id", placed.ID), zap.Int64("room_id", roomID),
				zap.Int("x", envelope.Command.X), zap.Int("y", envelope.Command.Y), zap.Int("rotation", envelope.Command.Rotation),
			)
		}
		return handler.publish(projectionCtx, player.ID(), placed.ID, item.DefinitionID, roomID, envelope.Command.X, envelope.Command.Y, envelope.Command.Rotation)
	})
}

// sendInventoryRemove notifies the actor that the placed item left their inventory.
func (handler Handler) sendInventoryRemove(ctx context.Context, connection netconn.Context, itemID int64) error {
	packet, err := outinvremove.Encode(itemID)
	if err != nil {
		return err
	}

	return connection.Send(ctx, packet)
}

// broadcastAdd notifies all room occupants that a floor item was placed.
func (handler Handler) broadcastAdd(ctx context.Context, active *roomlive.Room, item furnituremodel.Item, definition furnituremodel.Definition, ownerName string) error {
	if handler.Connections == nil {
		return nil
	}

	sender, err := handler.giftSender(ctx, item)
	if err != nil {
		return err
	}

	group, linked := furnituremodel.GroupData{}, false
	if handler.Groups != nil {
		group, linked = handler.Groups.Furniture(item.ID)
	}
	packet, err := outadd.Encode(addRecord(item, definition, ownerName, sender, group, linked))
	if err != nil {
		return err
	}

	return broadcast.RoomPacket(ctx, handler.Connections, active, packet, 0)
}

// broadcastHeightMapUpdate notifies all room occupants of the tile heights affected by the newly
// placed item's footprint, keeping every client's cached local height map (used for placement and
// movement prediction) in sync with the change.
func (handler Handler) broadcastHeightMapUpdate(ctx context.Context, active *roomlive.Room, placed worldfurniture.Item) error {
	if handler.Connections == nil {
		return nil
	}

	points := worldfurniture.Footprint(placed.Point, placed.Definition.Width, placed.Definition.Length, placed.Rotation)

	return broadcast.RoomHeightMapUpdate(ctx, handler.Connections, active, points, 0)
}

// publish emits furniture placement completion.
func (handler Handler) publish(ctx context.Context, playerID int64, itemID int64, definitionID int64, roomID int64, x int, y int, rotation int) error {
	if handler.Events == nil {
		return nil
	}

	return handler.Events.Publish(ctx, bus.Event{
		Name: placedevent.Name,
		Payload: placedevent.Payload{
			PlayerID: playerID, ItemID: itemID, DefinitionID: definitionID, RoomID: roomID, X: x, Y: y, Rotation: rotation,
		},
	})
}
