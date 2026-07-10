// Package list returns current room build rights.
package list

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	"github.com/niflaot/pixels/internal/realm/room/commands/control"
	roomrights "github.com/niflaot/pixels/internal/realm/room/rights"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	netconn "github.com/niflaot/pixels/networking/connection"
	outlist "github.com/niflaot/pixels/networking/outbound/room/rights/list"
)

const (
	// Name identifies the room rights list command.
	Name command.Name = "room.rights.list"
)

// Command requests room rights.
type Command struct {
	// Handler stores the source connection.
	Handler netconn.Context
	// RoomID identifies the packet room.
	RoomID int64
}

// Handler handles room rights lists.
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

// Handle sends current room rights.
func (handler Handler) Handle(ctx context.Context, envelope command.Envelope[Command]) error {
	_, roomID, err := control.Actor(envelope.Command.Handler, handler.Bindings, handler.Players)
	if err != nil {
		return err
	}
	if err := control.MatchRoom(roomID, envelope.Command.RoomID); err != nil {
		return err
	}
	rights, err := handler.Rights.ListRights(ctx, roomID)
	if err != nil {
		return err
	}
	records := make([]outlist.Right, len(rights))
	for index := range rights {
		records[index] = outlist.Right{PlayerID: int32(rights[index].PlayerID), Username: rights[index].Username}
	}
	packet, err := outlist.Encode(int32(roomID), records)
	if err != nil {
		return err
	}

	return envelope.Command.Handler.Send(ctx, packet)
}
