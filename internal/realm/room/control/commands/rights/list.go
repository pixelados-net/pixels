package rights

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	"github.com/niflaot/pixels/internal/realm/room/control/commands/resolve"
	roomrights "github.com/niflaot/pixels/internal/realm/room/control/rights"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	netconn "github.com/niflaot/pixels/networking/connection"
	outlist "github.com/niflaot/pixels/networking/outbound/room/rights/list"
)

const (
	// ListName identifies the room rights list command.
	ListName command.Name = "room.rights.list"
)

// ListCommand requests room rights.
type ListCommand struct {
	// Handler stores the source connection.
	Handler netconn.Context
	// RoomID identifies the packet room.
	RoomID int64
}

// ListHandler handles room rights lists.
type ListHandler struct {
	// Players stores live players.
	Players *playerlive.Registry
	// Bindings stores connection bindings.
	Bindings *binding.Registry
	// Rights manages room rights.
	Rights roomrights.Manager
}

// CommandName returns the stable command name.
func (ListCommand) CommandName() command.Name { return ListName }

// Handle sends current room rights.
func (handler ListHandler) Handle(ctx context.Context, envelope command.Envelope[ListCommand]) error {
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
