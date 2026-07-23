// Package stackheight applies exact custom stack-helper surface heights.
package stackheight

import (
	"context"
	"errors"

	permissionservice "github.com/niflaot/pixels/internal/permission/service"
	furnitureaccess "github.com/niflaot/pixels/internal/realm/furniture/access"
	furnituresession "github.com/niflaot/pixels/internal/realm/furniture/commands/session"
	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	"github.com/niflaot/pixels/internal/realm/room/runtime/broadcast"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	roomprojection "github.com/niflaot/pixels/internal/realm/room/runtime/projection"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	instack "github.com/niflaot/pixels/networking/inbound/furniture/stackheight"
	outstack "github.com/niflaot/pixels/networking/outbound/furniture/stackheight"
	outupdate "github.com/niflaot/pixels/networking/outbound/room/furniture/update"
)

const (
	// MaxHeightCM bounds the renderer slider and persisted surface height.
	MaxHeightCM int32 = 4000
)

// Handler applies stack-height changes for authorized room decorators.
type Handler struct {
	// Players stores live players.
	Players *playerlive.Registry
	// Bindings maps connections to authenticated players.
	Bindings *binding.Registry
	// Furniture reads and changes durable furniture.
	Furniture furnitureservice.Manager
	// States changes exact per-item height state.
	States furnitureservice.StackHeightUpdater
	// Runtime stores active rooms.
	Runtime *roomlive.Registry
	// Permissions resolves global furniture authority.
	Permissions permissionservice.Checker
	// Connections stores active room connections.
	Connections *netconn.Registry
}

// Register adds the stack-height request handler.
func Register(registry *netconn.HandlerRegistry, handler Handler) {
	if registry == nil {
		return
	}
	_ = registry.Register(instack.Header, handler.Handle)
}

// Handle decodes and applies one exact stack-height override.
func (handler Handler) Handle(connection netconn.Context, packet codec.Packet) error {
	payload, err := instack.Decode(packet)
	if err != nil {
		return err
	}
	return handler.apply(context.Background(), connection, payload)
}

// apply validates room ownership, persistence, and runtime projection.
func (handler Handler) apply(ctx context.Context, connection netconn.Context, payload instack.Payload) error {
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
	item, found, err := handler.Furniture.FindItemByID(ctx, int64(payload.ItemID))
	if err != nil || !found || item.RoomID == nil || *item.RoomID != roomID {
		return err
	}
	definition, found, err := handler.Furniture.FindDefinitionByID(ctx, item.DefinitionID)
	if err != nil || !found || definition.InteractionType != "custom_stack_height" {
		return err
	}
	height, err := normalizedHeight(payload.Height)
	if err != nil {
		return err
	}
	updated, err := handler.States.UpdateStackHeight(ctx, item.ID, roomID, height)
	if err != nil {
		return err
	}
	worldItem, found := active.FurnitureItem(item.ID)
	if !found {
		return nil
	}
	if height == nil {
		worldItem.Definition.StackHeight = worldItem.Definition.HeightAtState(worldItem.ExtraData)
	} else {
		worldItem.Definition.StackHeight = grid.HeightFromUnits(float64(*height) / 100)
	}
	units, err := active.ReloadFurniture(item.ID, &worldItem)
	if err != nil {
		return err
	}
	if err = broadcast.RoomUnitStatuses(ctx, handler.Connections, active, units, 0); err != nil {
		return err
	}
	if err = handler.broadcastUpdate(ctx, active, updated, definition); err != nil {
		return err
	}
	effective := payload.Height
	if height == nil {
		effective = int32(worldItem.Definition.StackHeight.Units() * 100)
	}
	response, err := outstack.Encode(payload.ItemID, effective)
	if err != nil {
		return err
	}
	return connection.Send(ctx, response)
}

// normalizedHeight maps the automatic sentinel and validates exact centimeters.
func normalizedHeight(height int32) (*int32, error) {
	if height == instack.AutoHeight {
		return nil, nil
	}
	if height < 0 || height > MaxHeightCM {
		return nil, errors.New("stack helper height out of range")
	}
	return &height, nil
}

// broadcastUpdate projects the helper's changed object snapshot.
func (handler Handler) broadcastUpdate(ctx context.Context, active *roomlive.Room, item furnituremodel.Item, definition furnituremodel.Definition) error {
	if item.X == nil || item.Y == nil || item.Z == nil {
		return nil
	}
	record := outupdate.FloorItem{ID: item.ID, SpriteID: definition.SpriteID, X: *item.X, Y: *item.Y, Rotation: int(item.Rotation), Z: roomprojection.FurnitureHeightValue(*item.Z), ExtraHeight: roomprojection.ExtraHeightValue(definition), ExtraData: item.ExtraData, OwnerID: item.OwnerPlayerID}
	record.Data = roomprojection.SpecializedObjectData(definition.InteractionType, item.ExtraData)
	packet, err := outupdate.Encode(record)
	if err != nil {
		return err
	}
	return broadcast.RoomPacket(ctx, handler.Connections, active, packet, 0)
}
