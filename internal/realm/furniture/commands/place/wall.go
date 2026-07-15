package place

import (
	"context"

	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	"github.com/niflaot/pixels/internal/realm/room/runtime/broadcast"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	outwalladd "github.com/niflaot/pixels/networking/outbound/room/furniture/walladd"
)

// placeWall persists and projects one inventory wall item.
func (handler Handler) placeWall(ctx context.Context, command Command, player *playerlive.Player, active *roomlive.Room, roomID int64, item furnituremodel.Item, definition furnituremodel.Definition) error {
	if !furnituremodel.ValidWallPosition(command.WallPosition) || definition.InteractionType == "postit" {
		return nil
	}
	uniqueInteraction := ""
	if definition.InteractionType == "dimmer" {
		uniqueInteraction = definition.InteractionType
		exists, err := handler.roomHasInteraction(ctx, roomID, uniqueInteraction)
		if err != nil {
			return err
		}
		if exists {
			return handler.sendBubbleAlert(ctx, command.Handler, "session.bubble.furniture.max_dimmers")
		}
	}
	placed, err := handler.Furniture.Place(ctx, furnitureservice.PlaceParams{
		ItemID: item.ID, ActorPlayerID: player.ID(), RoomID: roomID,
		WallPosition: command.WallPosition, UniqueInteractionType: uniqueInteraction,
	})
	if err != nil {
		return handler.handleSoftError(ctx, command, err)
	}
	if err = handler.sendInventoryRemove(ctx, command.Handler, placed.ID); err != nil {
		return err
	}
	packet, err := outwalladd.Encode(outwalladd.Item{
		ID: placed.ID, SpriteID: definition.SpriteID, WallPosition: command.WallPosition,
		ExtraData: placed.ExtraData, UsagePolicy: 0, OwnerID: placed.OwnerPlayerID, OwnerName: player.Username(),
	})
	if err != nil {
		return err
	}
	if err = broadcast.RoomPacket(ctx, handler.Connections, active, packet, 0); err != nil {
		return err
	}
	return handler.publish(ctx, player.ID(), placed.ID, item.DefinitionID, roomID, 0, 0, 0)
}

// roomHasInteraction reports whether one room already contains a specialized furniture type.
func (handler Handler) roomHasInteraction(ctx context.Context, roomID int64, interactionType string) (bool, error) {
	items, err := handler.Furniture.ListRoomItems(ctx, roomID)
	if err != nil {
		return false, err
	}
	for _, item := range items {
		definition, found, findErr := handler.Furniture.FindDefinitionByID(ctx, item.DefinitionID)
		if findErr != nil {
			return false, findErr
		}
		if found && definition.InteractionType == interactionType {
			return true, nil
		}
	}
	return false, nil
}
