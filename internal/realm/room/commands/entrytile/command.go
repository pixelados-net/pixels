// Package entrytile sends the current room entry tile settings.
package entrytile

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	entercommand "github.com/niflaot/pixels/internal/realm/room/commands/enter"
	roomsession "github.com/niflaot/pixels/internal/realm/room/commands/session"
	"github.com/niflaot/pixels/internal/realm/room/layout"
	roomservice "github.com/niflaot/pixels/internal/realm/room/service"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	netconn "github.com/niflaot/pixels/networking/connection"
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
	roomLayout, found, err := handler.Layouts.FindByName(ctx, room.ModelName)
	if err != nil {
		return err
	}
	if !found {
		return roomservice.ErrLayoutNotAvailable
	}

	return entercommand.SendEntryTile(ctx, envelope.Command.Handler, roomLayout)
}
