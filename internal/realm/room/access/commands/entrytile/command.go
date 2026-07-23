// Package entrytile sends the current room entry tile settings.
package entrytile

import (
	"context"
	"errors"

	"github.com/niflaot/pixels/internal/command"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	entercommand "github.com/niflaot/pixels/internal/realm/room/access/commands/enter"
	roomservice "github.com/niflaot/pixels/internal/realm/room/record/service"
	roomsession "github.com/niflaot/pixels/internal/realm/room/runtime/commands/session"
	"github.com/niflaot/pixels/internal/realm/room/world/layout"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	netconn "github.com/niflaot/pixels/networking/connection"
	outthickness "github.com/niflaot/pixels/networking/outbound/room/thickness/updated"
)

const (
	// Name identifies the room entry tile command.
	Name command.Name = "room.entrytile"
)

// Command sends the player's current room entry tile.
type Command struct {
	// Handler stores the source connection handler.
	Handler netconn.Context
}

// Handler handles room entry tile commands.
type Handler struct {
	// Players stores live player state.
	Players *playerlive.Registry

	// Bindings stores player connection bindings.
	Bindings *binding.Registry

	// Rooms reads room persistence.
	Rooms roomservice.Manager

	// Layouts reads room layouts.
	Layouts layout.Manager
}

// CommandName returns the stable command name.
func (Command) CommandName() command.Name {
	return Name
}

// Handle handles a room entry tile command.
func (handler Handler) Handle(ctx context.Context, envelope command.Envelope[Command]) error {
	player, err := roomsession.Player(envelope.Command.Handler, handler.Bindings, handler.Players)
	if err != nil {
		return err
	}
	roomID, found := player.CurrentRoom()
	if !found {
		return roomservice.ErrRoomNotFound
	}
	room, found, err := handler.Rooms.FindByID(ctx, roomID)
	if err != nil {
		return err
	}
	if !found {
		return roomservice.ErrRoomNotFound
	}
	roomLayout, err := layout.ResolveForRoom(ctx, handler.Layouts, room.ID, room.ModelName)
	if err != nil {
		if errors.Is(err, layout.ErrLayoutNotFound) {
			return roomservice.ErrLayoutNotAvailable
		}
		return err
	}

	if err = entercommand.SendEntryTile(ctx, envelope.Command.Handler, roomLayout); err != nil {
		return err
	}
	wall, floor := room.WallThickness, room.FloorThickness
	if roomLayout.RoomID > 0 {
		wall, floor = roomLayout.WallThickness, roomLayout.FloorThickness
	}
	packet, err := outthickness.Encode(room.HideWalls, int32(wall), int32(floor))
	if err != nil {
		return err
	}

	return envelope.Command.Handler.Send(ctx, packet)
}
