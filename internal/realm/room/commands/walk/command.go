// Package walk moves a live room unit toward a target tile.
package walk

import (
	"context"
	"errors"

	"github.com/niflaot/pixels/internal/command"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	"github.com/niflaot/pixels/internal/realm/room/broadcast"
	roomsession "github.com/niflaot/pixels/internal/realm/room/commands/session"
	roomlive "github.com/niflaot/pixels/internal/realm/room/live"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldpath "github.com/niflaot/pixels/internal/realm/room/world/path"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	netconn "github.com/niflaot/pixels/networking/connection"
)

const (
	// Name identifies the room walk command.
	Name command.Name = "room.walk"
)

var (
	// ErrPlayerNotInRoom reports a walk without active room presence.
	ErrPlayerNotInRoom = errors.New("player not in room")

	// ErrInvalidTarget reports malformed target coordinates.
	ErrInvalidTarget = errors.New("invalid walk target")
)

// Command moves a player unit.
type Command struct {
	// Handler stores the source connection handler.
	Handler netconn.Context

	// X stores the target tile x coordinate.
	X int

	// Y stores the target tile y coordinate.
	Y int
}

// Handler handles room walk commands.
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

// Handle handles a room walk command.
func (handler Handler) Handle(ctx context.Context, envelope command.Envelope[Command]) error {
	player, err := roomsession.Player(envelope.Command.Handler, handler.Bindings, handler.Players)
	if err != nil {
		return err
	}
	roomID, found := player.CurrentRoom()
	if !found {
		return ErrPlayerNotInRoom
	}
	active, found := handler.Runtime.Find(roomID)
	if !found {
		return roomlive.ErrRoomNotFound
	}
	point, ok := grid.NewPoint(envelope.Command.X, envelope.Command.Y)
	if !ok {
		return ErrInvalidTarget
	}

	if _, err := active.MoveTo(player.ID(), point); err != nil {
		return handler.handleMoveError(ctx, active, player.ID(), point, err)
	}

	return nil
}

// handleMoveError handles gameplay movement misses without closing the session.
func (handler Handler) handleMoveError(ctx context.Context, active *roomlive.Room, playerID int64, point grid.Point, err error) error {
	if !isSoftMoveError(err) {
		return err
	}
	unit, faceErr := active.FaceTo(playerID, point)
	if faceErr != nil {
		return nil
	}
	if handler.Connections == nil {
		return nil
	}

	return broadcast.RoomUnitStatus(ctx, handler.Connections, active, unit, 0)
}

// isSoftMoveError reports movement misses that should not disconnect clients. ErrInvalidStart in
// particular can surface when the surface underneath a unit changed height after it last moved (e.g.
// furniture it stood on was picked up or moved away); it is a stale-state miss, never a reason to
// drop the connection.
func isSoftMoveError(err error) bool {
	return errors.Is(err, worldpath.ErrInvalidGoal) ||
		errors.Is(err, worldpath.ErrNoPath) ||
		errors.Is(err, worldpath.ErrSearchLimit) ||
		errors.Is(err, worldpath.ErrInvalidPath) ||
		errors.Is(err, worldpath.ErrInvalidStart)
}
