package move

import (
	"context"

	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	"github.com/niflaot/pixels/internal/realm/room/runtime/broadcast"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/internal/realm/room/runtime/projection"
	outwallupdate "github.com/niflaot/pixels/networking/outbound/room/furniture/wallupdate"
)

// moveWall persists and projects one wall furniture reposition.
func (handler Handler) moveWall(ctx context.Context, command Command, player *playerlive.Player, active *roomlive.Room, roomID int64, item furnituremodel.Item) error {
	if !furnituremodel.ValidWallPosition(command.WallPosition) || item.RoomID == nil || *item.RoomID != roomID || item.WallPosition == nil {
		return nil
	}
	definition, found, err := handler.Furniture.FindDefinitionByID(ctx, item.DefinitionID)
	if err != nil || !found || definition.Kind != furnituremodel.KindWall {
		return err
	}
	moved, err := handler.Furniture.Move(ctx, furnitureservice.MoveParams{
		ItemID: item.ID, ActorPlayerID: player.ID(), RoomID: roomID, WallPosition: command.WallPosition,
	})
	if err != nil {
		return handler.handleSoftError(ctx, command, roomID, err)
	}
	packet, err := outwallupdate.Encode(moved.ID, definition.SpriteID, command.WallPosition, projection.WallExtraData(definition, moved.ExtraData), 0, moved.OwnerPlayerID)
	if err != nil {
		return err
	}
	if err = broadcast.RoomPacket(ctx, handler.Connections, active, packet, 0); err != nil {
		return err
	}
	return handler.publish(ctx, player.ID(), moved.ID, roomID, 0, 0, 0)
}
