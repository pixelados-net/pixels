package finals

import (
	"context"
	"encoding/json"

	furnitureaccess "github.com/niflaot/pixels/internal/realm/furniture/access"
	furnituresession "github.com/niflaot/pixels/internal/realm/furniture/commands/session"
	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
	"github.com/niflaot/pixels/internal/realm/room/runtime/broadcast"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	roomprojection "github.com/niflaot/pixels/internal/realm/room/runtime/projection"
	netconn "github.com/niflaot/pixels/networking/connection"
	inobjectdata "github.com/niflaot/pixels/networking/inbound/room/control/objectdata"
	outupdate "github.com/niflaot/pixels/networking/outbound/room/furniture/update"
)

// objectData changes one authorized custom-variable furniture instance.
func (handler Handler) objectData(ctx context.Context, connection netconn.Context, payload inobjectdata.Payload) error {
	player, err := furnituresession.Player(connection, handler.Bindings, handler.Players)
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
	if err != nil || !allowed {
		return err
	}
	item, found, err := handler.Furniture.FindItemByID(ctx, int64(payload.ObjectID))
	if err != nil || !found || item.RoomID == nil || *item.RoomID != roomID {
		return err
	}
	definition, found, err := handler.Furniture.FindDefinitionByID(ctx, item.DefinitionID)
	if err != nil || !found || definition.InteractionType != "custom_variables" {
		return err
	}
	encoded, err := json.Marshal(payload.Data)
	if err != nil {
		return err
	}
	updated, err := handler.States.UpdateState(ctx, furnitureservice.StateParams{ItemID: item.ID, RoomID: roomID, Expected: item.ExtraData, Next: string(encoded)})
	if err != nil {
		return err
	}
	active.SetFurnitureExtraData(updated.ID, updated.ExtraData)
	return handler.broadcastObjectData(ctx, active, updated, definition)
}

// broadcastObjectData projects map object data to current room occupants.
func (handler Handler) broadcastObjectData(ctx context.Context, active *roomlive.Room, item furnituremodel.Item, definition furnituremodel.Definition) error {
	if item.X == nil || item.Y == nil || item.Z == nil {
		return nil
	}
	record := outupdate.FloorItem{ID: item.ID, SpriteID: definition.SpriteID, X: *item.X, Y: *item.Y, Rotation: int(item.Rotation), Z: roomprojection.FurnitureHeightValue(*item.Z), ExtraHeight: roomprojection.ExtraHeightValue(definition), ExtraData: item.ExtraData, OwnerID: item.OwnerPlayerID}
	record.Data = roomprojection.SpecializedObjectData(definition.InteractionType, item.ExtraData)
	result, err := outupdate.Encode(record)
	if err != nil {
		return err
	}
	return broadcast.RoomPacket(ctx, handler.Connections, active, result, 0)
}
