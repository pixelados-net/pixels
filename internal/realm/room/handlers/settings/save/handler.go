// Package save adapts the room settings save packet.
package save

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	savecmd "github.com/niflaot/pixels/internal/realm/room/commands/settings/save"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	insave "github.com/niflaot/pixels/networking/inbound/room/settings/save"
	"go.uber.org/zap"
)

// New creates a room settings save packet handler.
func New(handler savecmd.Handler, log *zap.Logger) netconn.Handler {
	dispatcher, _ := command.NewDispatcher(handler)
	dispatcher.WithLogger(log)

	return func(connection netconn.Context, packet codec.Packet) error {
		payload, err := insave.Decode(packet)
		if err != nil {
			return err
		}

		return dispatcher.Dispatch(context.Background(), command.Envelope[savecmd.Command]{Command: mapCommand(connection, payload)})
	}
}

// mapCommand maps decoded packet fields to a settings command.
func mapCommand(connection netconn.Context, payload insave.Payload) savecmd.Command {
	return savecmd.Command{Handler: connection, RoomID: int64(payload.RoomID), Name: payload.Name,
		Description: payload.Description, DoorMode: payload.DoorMode, Password: payload.Password,
		MaxUsers: int(payload.MaxUsers), CategoryID: int64(payload.CategoryID), Tags: payload.Tags,
		TradeMode: payload.TradeMode, AllowPets: payload.AllowPets, AllowPetsEat: payload.AllowPetsEat,
		AllowWalkthrough: payload.AllowWalkthrough, HideWalls: payload.HideWalls,
		WallThickness: int(payload.WallThickness), FloorThickness: int(payload.FloorThickness),
		ModerationMute: payload.ModerationMute, ModerationKick: payload.ModerationKick, ModerationBan: payload.ModerationBan,
		ChatMode: int16(payload.ChatMode), ChatWeight: int16(payload.ChatWeight), ChatSpeed: int16(payload.ChatSpeed),
		ChatDistance: int16(payload.ChatDistance), ChatProtection: int16(payload.ChatProtection)}
}

// Register adds the room settings save handler.
func Register(registry *netconn.HandlerRegistry, handler netconn.Handler) {
	_ = registry.Register(insave.Header, handler)
}
