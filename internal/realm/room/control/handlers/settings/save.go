package settings

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	settingscmd "github.com/niflaot/pixels/internal/realm/room/control/commands/settings"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	insave "github.com/niflaot/pixels/networking/inbound/room/settings/save"
	"go.uber.org/zap"
)

// NewSave creates a room settings save packet handler.
func NewSave(handler settingscmd.SaveHandler, log *zap.Logger) netconn.Handler {
	dispatcher, _ := command.NewDispatcher(handler)
	dispatcher.WithLogger(log)

	return func(connection netconn.Context, packet codec.Packet) error {
		payload, err := insave.Decode(packet)
		if err != nil {
			return err
		}

		return dispatcher.Dispatch(context.Background(), command.Envelope[settingscmd.SaveCommand]{Command: mapCommand(connection, payload)})
	}
}

// mapCommand maps decoded packet fields to a settings command.
func mapCommand(connection netconn.Context, payload insave.Payload) settingscmd.SaveCommand {
	return settingscmd.SaveCommand{Handler: connection, RoomID: int64(payload.RoomID), Name: payload.Name,
		Description: payload.Description, DoorMode: payload.DoorMode, Password: payload.Password,
		MaxUsers: int(payload.MaxUsers), CategoryID: int64(payload.CategoryID), Tags: payload.Tags,
		TradeMode: payload.TradeMode, AllowPets: payload.AllowPets, AllowPetsEat: payload.AllowPetsEat,
		AllowWalkthrough: payload.AllowWalkthrough, HideWalls: payload.HideWalls,
		WallThickness: int(payload.WallThickness), FloorThickness: int(payload.FloorThickness),
		ModerationMute: payload.ModerationMute, ModerationKick: payload.ModerationKick, ModerationBan: payload.ModerationBan,
		ChatMode: int16(payload.ChatMode), ChatWeight: int16(payload.ChatWeight), ChatSpeed: int16(payload.ChatSpeed),
		ChatDistance: int16(payload.ChatDistance), ChatProtection: int16(payload.ChatProtection)}
}

// RegisterSave adds the room settings save handler.
func RegisterSave(registry *netconn.HandlerRegistry, handler netconn.Handler) {
	_ = registry.Register(insave.Header, handler)
}
