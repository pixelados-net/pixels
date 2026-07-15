// Package pickup returns a placed furniture item to its owner's inventory.
package pickup

import (
	"context"
	"errors"

	"github.com/niflaot/pixels/internal/command"
	permissionservice "github.com/niflaot/pixels/internal/permission/service"
	furnitureaccess "github.com/niflaot/pixels/internal/realm/furniture/access"
	furnituresession "github.com/niflaot/pixels/internal/realm/furniture/commands/session"
	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	"github.com/niflaot/pixels/internal/realm/room/runtime/broadcast"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	worldfurniture "github.com/niflaot/pixels/internal/realm/room/world/furniture"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	outrefresh "github.com/niflaot/pixels/networking/outbound/inventory/furniture/refresh"
	outunseen "github.com/niflaot/pixels/networking/outbound/inventory/unseen"
	outremove "github.com/niflaot/pixels/networking/outbound/room/furniture/remove"
	outwallremove "github.com/niflaot/pixels/networking/outbound/room/furniture/wallremove"
	"github.com/niflaot/pixels/pkg/bus"
	"github.com/niflaot/pixels/pkg/i18n"
	"go.uber.org/zap"
)

const (
	// Name identifies the furniture pickup command.
	Name command.Name = "furniture.pickup"

	// bubbleKeyFurniturePlacementError matches Arcturus's BubbleAlertKeys.FURNITURE_PLACEMENT_ERROR.
	bubbleKeyFurniturePlacementError = "furni_placement_error"
)

// ErrPlayerNotInRoom reports a pickup attempt without active room presence.
var ErrPlayerNotInRoom = errors.New("player not in room")

// Command returns a placed furniture item to inventory.
type Command struct {
	// Handler stores the source connection handler.
	Handler netconn.Context

	// ItemID identifies the placed furniture item to pick up.
	ItemID int64
}

// Handler handles furniture pickup commands.
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

	// Log records rejected pickup attempts.
	Log *zap.Logger
}

// CommandName returns the stable command name.
func (Command) CommandName() command.Name {
	return Name
}

// Handle handles a furniture pickup command.
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

	picked, err := handler.Furniture.Pickup(ctx, furnitureservice.PickupParams{
		ItemID:        envelope.Command.ItemID,
		ActorPlayerID: player.ID(),
		RoomID:        roomID,
		AllowForeign:  true,
	})
	if err != nil {
		return handler.handleSoftError(ctx, envelope.Command, err)
	}

	stoodUp, err := active.ReloadFurniture(picked.ID, nil)
	if err != nil {
		return err
	}
	if err := handler.broadcastRemove(ctx, active, picked); err != nil {
		return err
	}
	if err := handler.broadcastStoodUpOccupants(ctx, active, stoodUp); err != nil {
		return err
	}
	if err := handler.broadcastHeightMapUpdate(ctx, active, picked); err != nil {
		return err
	}
	if err := handler.sendInventoryUpdate(ctx, envelope.Command.Handler, picked.ID); err != nil {
		return err
	}

	if handler.Log != nil {
		handler.Log.Debug("furniture picked up",
			zap.Int64("player_id", player.ID()), zap.Int64("item_id", picked.ID), zap.Int64("room_id", roomID),
		)
	}

	return handler.publish(ctx, player.ID(), picked.ID, roomID)
}

// broadcastRemove notifies all room occupants that a floor item was picked up.
func (handler Handler) broadcastRemove(ctx context.Context, active *roomlive.Room, item furnituremodel.Item) error {
	if handler.Connections == nil {
		return nil
	}
	definition, found, err := handler.Furniture.FindDefinitionByID(ctx, item.DefinitionID)
	if err != nil {
		return err
	}
	var packet codec.Packet
	if found && definition.Kind == furnituremodel.KindWall {
		packet, err = outwallremove.Encode(item.ID, item.OwnerPlayerID)
	} else {
		packet, err = outremove.Encode(item.ID, item.OwnerPlayerID)
	}
	if err != nil {
		return err
	}

	return broadcast.RoomPacket(ctx, handler.Connections, active, packet, 0)
}

// broadcastStoodUpOccupants notifies all room occupants of units the picked up item's occupant(s)
// stood up (see Room.ReloadFurniture).
func (handler Handler) broadcastStoodUpOccupants(ctx context.Context, active *roomlive.Room, units []roomlive.UnitSnapshot) error {
	if handler.Connections == nil {
		return nil
	}

	return broadcast.RoomUnitStatuses(ctx, handler.Connections, active, units, 0)
}

// broadcastHeightMapUpdate notifies all room occupants of the tile heights freed by the picked up
// item's old footprint, keeping every client's cached local height map (used for placement and
// movement prediction) in sync with the change.
func (handler Handler) broadcastHeightMapUpdate(ctx context.Context, active *roomlive.Room, picked furnituremodel.Item) error {
	if handler.Connections == nil || picked.X == nil || picked.Y == nil {
		return nil
	}
	point, ok := grid.NewPoint(*picked.X, *picked.Y)
	if !ok {
		return nil
	}

	definition, found, err := handler.Furniture.FindDefinitionByID(ctx, picked.DefinitionID)
	if err != nil {
		return err
	}
	if !found {
		return nil
	}

	points := worldfurniture.Footprint(point, definition.Width, definition.Length, worldunit.Rotation(picked.Rotation))

	return broadcast.RoomHeightMapUpdate(ctx, handler.Connections, active, points, 0)
}

// sendInventoryUpdate marks a picked up item unseen and invalidates inventory data.
func (handler Handler) sendInventoryUpdate(ctx context.Context, connection netconn.Context, itemID int64) error {
	packet, err := outunseen.EncodeOwned([]int64{itemID})
	if err != nil {
		return err
	}
	if err := connection.Send(ctx, packet); err != nil {
		return err
	}
	packet, err = outrefresh.Encode()
	if err != nil {
		return err
	}

	return connection.Send(ctx, packet)
}
