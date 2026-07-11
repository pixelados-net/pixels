// Package look rotates a room unit toward a target tile.
package look

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	"github.com/niflaot/pixels/internal/realm/room/runtime/broadcast"
	roomsession "github.com/niflaot/pixels/internal/realm/room/runtime/commands/session"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	netconn "github.com/niflaot/pixels/networking/connection"
)

const (
	// Name identifies the room look command.
	Name command.Name = "room.look"
)

// Command rotates the player's active room unit.
type Command struct {
	// Handler stores the source connection handler.
	Handler netconn.Context

	// X stores the target tile x coordinate.
	X int

	// Y stores the target tile y coordinate.
	Y int
}

// Handler handles room look commands.
type Handler struct {
	// Players stores live player state.
	Players *playerlive.Registry

	// Bindings stores player connection bindings.
	Bindings *binding.Registry

	// Runtime stores active rooms.
	Runtime *roomlive.Registry

	// Connections stores active network connections.
	Connections *netconn.Registry
}

// CommandName returns the stable command name.
func (Command) CommandName() command.Name {
	return Name
}

// Handle handles a room look command.
func (handler Handler) Handle(ctx context.Context, envelope command.Envelope[Command]) error {
	player, err := roomsession.Player(envelope.Command.Handler, handler.Bindings, handler.Players)
	if err != nil {
		return err
	}
	roomID, found := player.CurrentRoom()
	if !found {
		return ErrPlayerNotInRoom
	}
	point, valid := grid.NewPoint(envelope.Command.X, envelope.Command.Y)
	if !valid {
		return ErrInvalidTarget
	}
	active, found := handler.Runtime.Find(roomID)
	if !found {
		return roomlive.ErrRoomNotFound
	}
	if movingUnit(active, player.ID()) {
		return nil
	}

	unit, err := active.FaceTo(player.ID(), point)
	if err != nil {
		return err
	}
	if handler.Connections == nil {
		return nil
	}

	return broadcast.RoomUnitStatus(ctx, handler.Connections, active, unit, 0)
}

// movingUnit reports whether a player unit is currently walking.
func movingUnit(active *roomlive.Room, playerID int64) bool {
	for _, unit := range active.Units() {
		if unit.PlayerID == playerID {
			return unit.Moving
		}
	}

	return false
}
