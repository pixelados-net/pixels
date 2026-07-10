// Package revokeall revokes every room build-right holder.
package revokeall

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	"github.com/niflaot/pixels/internal/realm/room/commands/control"
	roomrights "github.com/niflaot/pixels/internal/realm/room/rights"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	netconn "github.com/niflaot/pixels/networking/connection"
)

const (
	// Name identifies the room rights revoke-all command.
	Name command.Name = "room.rights.revoke_all"
)

// Command revokes every room rights holder.
type Command struct {
	// Handler stores the source connection.
	Handler netconn.Context
	// RoomID identifies the packet room.
	RoomID int64
}

// Handler handles revoke-all requests.
type Handler struct {
	// Players stores live players.
	Players *playerlive.Registry
	// Bindings stores connection bindings.
	Bindings *binding.Registry
	// Rights manages room rights.
	Rights roomrights.Manager
}

// CommandName returns the stable command name.
func (Command) CommandName() command.Name { return Name }

// Handle revokes all rights in the current room.
func (handler Handler) Handle(ctx context.Context, envelope command.Envelope[Command]) error {
	player, roomID, err := control.Actor(envelope.Command.Handler, handler.Bindings, handler.Players)
	if err != nil {
		return err
	}
	if err := control.MatchRoom(roomID, envelope.Command.RoomID); err != nil {
		return err
	}
	_, err = handler.Rights.RevokeAllRights(ctx, roomID, player.ID())

	return err
}
