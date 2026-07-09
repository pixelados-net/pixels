// Package place places an inventory furniture item into the player's current room.
package place

import (
	"context"
	"errors"

	"github.com/niflaot/pixels/internal/command"
	furnituresession "github.com/niflaot/pixels/internal/realm/furniture/commands/session"
	placedevent "github.com/niflaot/pixels/internal/realm/furniture/events/placed"
	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	"github.com/niflaot/pixels/internal/realm/room/broadcast"
	roomfurniture "github.com/niflaot/pixels/internal/realm/room/furniture"
	roomlive "github.com/niflaot/pixels/internal/realm/room/live"
	"github.com/niflaot/pixels/internal/realm/room/projection"
	worldfurniture "github.com/niflaot/pixels/internal/realm/room/world/furniture"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	netconn "github.com/niflaot/pixels/networking/connection"
	outinvremove "github.com/niflaot/pixels/networking/outbound/inventory/furniture/remove"
	outadd "github.com/niflaot/pixels/networking/outbound/room/furniture/add"
	outbubble "github.com/niflaot/pixels/networking/outbound/session/bubblealert"
	"github.com/niflaot/pixels/pkg/bus"
	"go.uber.org/zap"
)

const (
	// Name identifies the furniture place command.
	Name command.Name = "furniture.place"

	// bubbleKeyFurniturePlacementError matches Arcturus's BubbleAlertKeys.FURNITURE_PLACEMENT_ERROR.
	bubbleKeyFurniturePlacementError = "furni_placement_error"
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
}

// Handler handles furniture place commands.
type Handler struct {
	// Players stores live player state.
	Players *playerlive.Registry

	// Bindings stores player connection bindings.
	Bindings *binding.Registry

	// Furniture manages placed and inventory furniture records.
	Furniture furnitureservice.Manager

	// Runtime stores active rooms.
	Runtime *roomlive.Registry

	// Connections stores active network connections.
	Connections *netconn.Registry

	// Events publishes furniture lifecycle events.
	Events bus.Publisher

	// Log records rejected placement attempts.
	Log *zap.Logger
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

	item, found, err := handler.Furniture.FindItemByID(ctx, envelope.Command.ItemID)
	if err != nil {
		return err
	}
	if !found {
		return nil
	}

	rotation := furnituremodel.Rotation(envelope.Command.Rotation)
	worldItem, definition, err := roomfurniture.ResolveWorldItem(ctx, active, handler.Furniture, item.ID, item.DefinitionID, envelope.Command.X, envelope.Command.Y, rotation)
	if err != nil {
		return handler.handleSoftError(ctx, envelope.Command, err)
	}

	placed, err := handler.Furniture.Place(ctx, furnitureservice.PlaceParams{
		ItemID:        item.ID,
		ActorPlayerID: player.ID(),
		RoomID:        roomID,
		Placement: furnituremodel.Placement{
			X: envelope.Command.X, Y: envelope.Command.Y,
			Z: float64(worldItem.Z), Rotation: rotation,
		},
	})
	if err != nil {
		return handler.handleSoftError(ctx, envelope.Command, err)
	}

	if _, err := active.ReloadFurniture(placed.ID, &worldItem); err != nil {
		return err
	}
	if err := handler.sendInventoryRemove(ctx, envelope.Command.Handler, placed.ID); err != nil {
		return err
	}
	if err := handler.broadcastAdd(ctx, active, placed, definition, player.Username()); err != nil {
		return err
	}
	if err := handler.broadcastHeightMapUpdate(ctx, active, worldItem); err != nil {
		return err
	}

	if handler.Log != nil {
		handler.Log.Debug("furniture placed",
			zap.Int64("player_id", player.ID()), zap.Int64("item_id", placed.ID), zap.Int64("room_id", roomID),
			zap.Int("x", envelope.Command.X), zap.Int("y", envelope.Command.Y), zap.Int("rotation", envelope.Command.Rotation),
		)
	}

	return handler.publish(ctx, player.ID(), placed.ID, item.DefinitionID, roomID, envelope.Command.X, envelope.Command.Y, envelope.Command.Rotation)
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

	packet, err := outadd.Encode(addRecord(item, definition, ownerName))
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

// addRecord maps a placed item and its definition into an ADD_FLOOR_ITEM record.
func addRecord(item furnituremodel.Item, definition furnituremodel.Definition, ownerName string) outadd.FloorItem {
	return outadd.FloorItem{
		ID:          item.ID,
		SpriteID:    definition.SpriteID,
		X:           *item.X,
		Y:           *item.Y,
		Rotation:    int(item.Rotation),
		Z:           projection.FurnitureHeightValue(*item.Z),
		ExtraHeight: projection.ExtraHeightValue(definition),
		ExtraData:   item.ExtraData,
		UsagePolicy: projection.UsagePolicyValue(definition),
		OwnerID:     item.OwnerPlayerID,
		OwnerName:   ownerName,
	}
}

// handleSoftError logs a rejected placement attempt with context and sends a bubble alert when the
// error maps to a client-facing key, swallowing the error so the client is not disconnected.
func (handler Handler) handleSoftError(ctx context.Context, cmd Command, err error) error {
	key, soft := bubbleErrorKey(err)
	if !soft {
		return err
	}

	if handler.Log != nil {
		handler.Log.Warn("furniture placement rejected",
			zap.Int64("item_id", cmd.ItemID), zap.Int("x", cmd.X), zap.Int("y", cmd.Y), zap.Int("rotation", cmd.Rotation),
			zap.Error(err),
		)
	}
	if key == "" {
		return nil
	}

	return handler.sendBubbleAlert(ctx, cmd.Handler, key)
}

// sendBubbleAlert notifies the actor of a rejected furniture placement.
func (handler Handler) sendBubbleAlert(ctx context.Context, connection netconn.Context, key string) error {
	packet, err := outbubble.Encode(bubbleKeyFurniturePlacementError, key)
	if err != nil {
		return err
	}

	return connection.Send(ctx, packet)
}

// bubbleErrorKey reports whether an error is a soft gameplay miss and its bubble alert key, if any.
func bubbleErrorKey(err error) (string, bool) {
	switch {
	case errors.Is(err, roomlive.ErrInvalidPlacement),
		errors.Is(err, roomfurniture.ErrInvalidTarget),
		errors.Is(err, furnitureservice.ErrInvalidPlacement),
		errors.Is(err, furnitureservice.ErrInvalidItemID),
		errors.Is(err, furnitureservice.ErrInvalidRoomID),
		errors.Is(err, furnitureservice.ErrInvalidPlayerID):
		return "invalid_move", true
	case errors.Is(err, roomlive.ErrTileOccupied):
		return "tile_has_units", true
	case errors.Is(err, roomlive.ErrCannotStack):
		return "cant_stack", true
	case errors.Is(err, furnitureservice.ErrNotItemOwner):
		return "no_rights", true
	case errors.Is(err, furnitureservice.ErrItemNotInInventory):
		return "item_not_in_inventory", true
	case errors.Is(err, furnitureservice.ErrItemNotFound),
		errors.Is(err, furnitureservice.ErrItemNotPlaced),
		errors.Is(err, roomfurniture.ErrDefinitionNotFound):
		return "item_not_found", true
	case errors.Is(err, roomlive.ErrWorldNotLoaded):
		return "", true
	default:
		return "", false
	}
}
