// Package move repositions or rotates an already placed furniture item.
package move

import (
	"context"
	"errors"

	"github.com/niflaot/pixels/internal/command"
	permissionservice "github.com/niflaot/pixels/internal/permission/service"
	furnitureaccess "github.com/niflaot/pixels/internal/realm/furniture/access"
	furnituresession "github.com/niflaot/pixels/internal/realm/furniture/commands/session"
	movedevent "github.com/niflaot/pixels/internal/realm/furniture/events/moved"
	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	"github.com/niflaot/pixels/internal/realm/room/runtime/broadcast"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/internal/realm/room/runtime/projection"
	worldfurniture "github.com/niflaot/pixels/internal/realm/room/world/furniture"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	roomfurniture "github.com/niflaot/pixels/internal/realm/room/world/items"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	netconn "github.com/niflaot/pixels/networking/connection"
	outupdate "github.com/niflaot/pixels/networking/outbound/room/furniture/update"
	"github.com/niflaot/pixels/pkg/bus"
	"github.com/niflaot/pixels/pkg/i18n"
	"go.uber.org/zap"
)

const (
	// Name identifies the furniture move command.
	Name command.Name = "furniture.move"
)

// ErrPlayerNotInRoom reports a move attempt without active room presence.
var ErrPlayerNotInRoom = errors.New("player not in room")

// Command repositions or rotates a placed furniture item.
type Command struct {
	// Handler stores the source connection handler.
	Handler netconn.Context

	// ItemID identifies the placed furniture item to reposition.
	ItemID int64

	// X stores the destination floor tile x coordinate.
	X int

	// Y stores the destination floor tile y coordinate.
	Y int

	// Rotation stores the destination floor instance rotation.
	Rotation int
}

// Handler handles furniture move commands.
type Handler struct {
	// Players stores live player state.
	Players *playerlive.Registry

	// Bindings stores player connection bindings.
	Bindings *binding.Registry

	// Furniture manages placed and inventory furniture records.
	Furniture furnitureservice.Manager

	// Runtime stores active rooms.
	Runtime *roomlive.Registry

	// Permissions resolves global furniture management authority.
	Permissions permissionservice.Checker

	// Connections stores active network connections.
	Connections *netconn.Registry

	// Events publishes furniture lifecycle events.
	Events bus.Publisher

	// Translations resolves end-user messages.
	Translations i18n.Translator

	// Log records rejected move attempts.
	Log *zap.Logger
}

// CommandName returns the stable command name.
func (Command) CommandName() command.Name {
	return Name
}

// Handle handles a furniture move command.
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
		return handler.handleSoftError(ctx, envelope.Command, roomID, roomlive.ErrNoFurnitureRights)
	}

	item, found, err := handler.Furniture.FindItemByID(ctx, envelope.Command.ItemID)
	if err != nil {
		return err
	}
	if !found {
		return nil
	}

	rotation := furnituremodel.Rotation(envelope.Command.Rotation)
	worldItem, definition, err := roomfurniture.ResolveWorldItem(ctx, active, handler.Furniture, item, envelope.Command.X, envelope.Command.Y, rotation)
	if err != nil {
		return handler.handleSoftError(ctx, envelope.Command, roomID, err)
	}
	previousWorld, previousFound := active.FurnitureItem(item.ID)

	moved, err := handler.Furniture.Move(ctx, furnitureservice.MoveParams{
		ItemID:        item.ID,
		ActorPlayerID: player.ID(),
		RoomID:        roomID,
		Placement: furnituremodel.Placement{
			X: envelope.Command.X, Y: envelope.Command.Y,
			Z: float64(worldItem.Z), Rotation: rotation,
		},
	})
	if err != nil {
		return handler.handleSoftError(ctx, envelope.Command, roomID, err)
	}

	reoriented, err := active.ReloadFurniture(moved.ID, &worldItem)
	if err != nil {
		return err
	}
	if err := handler.broadcastUpdate(ctx, active, moved, definition); err != nil {
		return err
	}
	if err := handler.broadcastReorientedOccupants(ctx, active, reoriented); err != nil {
		return err
	}
	if err := handler.broadcastHeightMapUpdate(ctx, active, item, definition, worldItem); err != nil {
		return err
	}

	if handler.Log != nil {
		handler.Log.Debug("furniture moved",
			zap.Int64("player_id", player.ID()), zap.Int64("item_id", moved.ID), zap.Int64("room_id", roomID),
			zap.Int("x", envelope.Command.X), zap.Int("y", envelope.Command.Y), zap.Int("rotation", envelope.Command.Rotation),
		)
	}

	return errors.Join(
		handler.publishWalkedOff(ctx, active, previousWorld, previousFound, worldItem),
		handler.publish(ctx, player.ID(), moved.ID, roomID, envelope.Command.X, envelope.Command.Y, envelope.Command.Rotation),
	)
}

// broadcastUpdate notifies all room occupants that a floor item moved, rotated, or changed state.
func (handler Handler) broadcastUpdate(ctx context.Context, active *roomlive.Room, item furnituremodel.Item, definition furnituremodel.Definition) error {
	if handler.Connections == nil {
		return nil
	}

	packet, err := outupdate.Encode(updateRecord(item, definition))
	if err != nil {
		return err
	}

	return broadcast.RoomPacket(ctx, handler.Connections, active, packet, 0)
}

// broadcastReorientedOccupants notifies all room occupants of units the moved or rotated item
// re-settled or stood up (see Room.ReloadFurniture).
func (handler Handler) broadcastReorientedOccupants(ctx context.Context, active *roomlive.Room, units []roomlive.UnitSnapshot) error {
	if handler.Connections == nil {
		return nil
	}

	return broadcast.RoomUnitStatuses(ctx, handler.Connections, active, units, 0)
}

// broadcastHeightMapUpdate notifies all room occupants of the tile heights affected by a moved or
// rotated item, covering both its old and new footprint so every client's cached local height map
// (used for placement and movement prediction) stays in sync with the change.
func (handler Handler) broadcastHeightMapUpdate(ctx context.Context, active *roomlive.Room, previous furnituremodel.Item, definition furnituremodel.Definition, moved worldfurniture.Item) error {
	if handler.Connections == nil {
		return nil
	}

	points := make([]grid.Point, 0, 8)
	if previous.X != nil && previous.Y != nil {
		if oldPoint, ok := grid.NewPoint(*previous.X, *previous.Y); ok {
			points = append(points, worldfurniture.Footprint(oldPoint, definition.Width, definition.Length, worldunit.Rotation(previous.Rotation))...)
		}
	}
	points = append(points, worldfurniture.Footprint(moved.Point, moved.Definition.Width, moved.Definition.Length, moved.Rotation)...)

	return broadcast.RoomHeightMapUpdate(ctx, handler.Connections, active, points, 0)
}

// publish emits furniture move completion.
func (handler Handler) publish(ctx context.Context, playerID int64, itemID int64, roomID int64, x int, y int, rotation int) error {
	if handler.Events == nil {
		return nil
	}

	return handler.Events.Publish(ctx, bus.Event{
		Name: movedevent.Name,
		Payload: movedevent.Payload{
			PlayerID: playerID, ItemID: itemID, RoomID: roomID, X: x, Y: y, Rotation: rotation,
		},
	})
}

// updateRecord maps a moved item and its definition into a FLOOR_ITEM_UPDATE record.
func updateRecord(item furnituremodel.Item, definition furnituremodel.Definition) outupdate.FloorItem {
	return outupdate.FloorItem{
		ID:          item.ID,
		SpriteID:    definition.SpriteID,
		X:           *item.X,
		Y:           *item.Y,
		Rotation:    int(item.Rotation),
		Z:           projection.FurnitureHeightValue(*item.Z),
		ExtraHeight: projection.ExtraHeightValue(definition),
		ExtraData:   item.ExtraData,
		OwnerID:     item.OwnerPlayerID,
	}
}
