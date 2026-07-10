// Package listbans returns active room bans.
package listbans

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	"github.com/niflaot/pixels/internal/realm/room/commands/control"
	roommoderation "github.com/niflaot/pixels/internal/realm/room/moderation"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	netconn "github.com/niflaot/pixels/networking/connection"
	outbanlist "github.com/niflaot/pixels/networking/outbound/room/moderation/banlist"
)

const (
	// Name identifies the room ban-list command.
	Name command.Name = "room.moderation.list_bans"
)

// Command requests active room bans.
type Command struct {
	// Handler stores the source connection.
	Handler netconn.Context
	// RoomID identifies the packet room.
	RoomID int64
}

// Handler handles room ban-list commands.
type Handler struct {
	// Players stores live players.
	Players *playerlive.Registry
	// Bindings stores connection bindings.
	Bindings *binding.Registry
	// Moderation reads room moderation.
	Moderation roommoderation.Reader
}

// CommandName returns the stable command name.
func (Command) CommandName() command.Name { return Name }

// Handle sends active room bans.
func (handler Handler) Handle(ctx context.Context, envelope command.Envelope[Command]) error {
	_, roomID, err := control.Actor(envelope.Command.Handler, handler.Bindings, handler.Players)
	if err != nil {
		return err
	}
	if err := control.MatchRoom(roomID, envelope.Command.RoomID); err != nil {
		return err
	}
	bans, err := handler.Moderation.ListBans(ctx, roomID)
	if err != nil {
		return err
	}
	records := make([]outbanlist.Ban, len(bans))
	for index := range bans {
		records[index] = outbanlist.Ban{PlayerID: int32(bans[index].PlayerID), Username: bans[index].Username}
	}
	packet, err := outbanlist.Encode(int32(roomID), records)
	if err != nil {
		return err
	}

	return envelope.Command.Handler.Send(ctx, packet)
}
