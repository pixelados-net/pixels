// Package votes contains room vote commands.
package votes

import (
	"context"
	"errors"

	"github.com/niflaot/pixels/internal/command"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	"github.com/niflaot/pixels/internal/realm/room/control/commands/resolve"
	roomvotes "github.com/niflaot/pixels/internal/realm/room/control/votes"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	netconn "github.com/niflaot/pixels/networking/connection"
)

const (
	// CastName identifies the room vote command.
	CastName command.Name = "room.vote.cast"
	// PositiveRating stores Nitro's supported upvote value.
	PositiveRating int32 = 1
)

var (
	// ErrInvalidRating reports unsupported room rating values.
	ErrInvalidRating = errors.New("invalid room rating value")
)

// CastCommand casts one permanent room upvote.
type CastCommand struct {
	// Handler stores the source connection.
	Handler netconn.Context
	// Rating stores Nitro's requested positive rating.
	Rating int32
}

// CastHandler handles room vote commands.
type CastHandler struct {
	// Players stores live players.
	Players *playerlive.Registry
	// Bindings stores connection bindings.
	Bindings *binding.Registry
	// Votes manages durable room votes.
	Votes roomvotes.Manager
}

// CommandName returns the stable command name.
func (CastCommand) CommandName() command.Name { return CastName }

// Handle casts a vote for the actor's current room.
func (handler CastHandler) Handle(ctx context.Context, envelope command.Envelope[CastCommand]) error {
	if envelope.Command.Rating != PositiveRating {
		return ErrInvalidRating
	}
	player, roomID, err := control.Actor(envelope.Command.Handler, handler.Bindings, handler.Players)
	if err != nil {
		return err
	}

	_, err = handler.Votes.Cast(ctx, roomID, player.ID())
	return err
}
