// Package use handles furniture teleport use commands.
package use

import (
	"context"
	"errors"

	"github.com/niflaot/pixels/internal/command"
	furnituresession "github.com/niflaot/pixels/internal/realm/furniture/commands/session"
	teleport "github.com/niflaot/pixels/internal/realm/furniture/interactions/teleport"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	netconn "github.com/niflaot/pixels/networking/connection"
)

const (
	// Name identifies the furniture teleport use command.
	Name command.Name = "furniture.teleport.use"
)

// Command requests use of one paired teleport item.
type Command struct {
	// Handler stores the source connection handler.
	Handler netconn.Context
	// ItemID identifies the clicked furniture item.
	ItemID int64
	// State stores the client interaction state parameter.
	State int32
}

// Handler handles furniture teleport use commands.
type Handler struct {
	// Players stores live player state.
	Players *playerlive.Registry
	// Bindings stores player connection bindings.
	Bindings *binding.Registry
	// Runtime stores active rooms.
	Runtime *roomlive.Registry
	// Teleports coordinates teleport transitions.
	Teleports *teleport.Service
}

// CommandName returns the stable command name.
func (Command) CommandName() command.Name {
	return Name
}

// Handle starts a paired teleport transition.
func (handler Handler) Handle(ctx context.Context, envelope command.Envelope[Command]) error {
	player, err := furnituresession.Player(envelope.Command.Handler, handler.Bindings, handler.Players)
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
	err = handler.Teleports.Start(ctx, teleport.StartRequest{
		PlayerID: player.ID(), Room: active, ItemID: envelope.Command.ItemID,
	})
	if errors.Is(err, teleport.ErrNotTeleport) || errors.Is(err, teleport.ErrInvalidUse) {
		return nil
	}

	return err
}
