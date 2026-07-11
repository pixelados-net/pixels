package moderation

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	"github.com/niflaot/pixels/internal/realm/room/control/commands/resolve"
	roommoderation "github.com/niflaot/pixels/internal/realm/room/control/moderation"
	roomsettings "github.com/niflaot/pixels/internal/realm/room/control/settings"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	netconn "github.com/niflaot/pixels/networking/connection"
	outbanlist "github.com/niflaot/pixels/networking/outbound/room/moderation/banlist"
)

const (
	// BanListName identifies the room ban-list command.
	BanListName command.Name = "room.moderation.list_bans"
)

// BanListCommand requests active room bans.
type BanListCommand struct {
	// Handler stores the source connection.
	Handler netconn.Context
	// RoomID identifies the packet room.
	RoomID int64
}

// BanListHandler handles room ban-list commands.
type BanListHandler struct {
	// Players stores live players.
	Players *playerlive.Registry
	// Bindings stores connection bindings.
	Bindings *binding.Registry
	// Moderation reads room moderation.
	Moderation roommoderation.Reader
	// Rooms reads room ownership.
	Rooms roommoderation.RoomFinder
	// Authorize protects moderation configuration reads.
	Authorize *roomsettings.Authorizer
}

// CommandName returns the stable command name.
func (BanListCommand) CommandName() command.Name { return BanListName }

// Handle sends active room bans.
func (handler BanListHandler) Handle(ctx context.Context, envelope command.Envelope[BanListCommand]) error {
	player, roomID, err := control.Actor(envelope.Command.Handler, handler.Bindings, handler.Players)
	if err != nil {
		return err
	}
	if err := control.MatchRoom(roomID, envelope.Command.RoomID); err != nil {
		return err
	}
	room, found, err := handler.Rooms.FindByID(ctx, roomID)
	if err != nil {
		return err
	}
	if !found {
		return roommoderation.ErrRoomNotFound
	}
	if err := handler.Authorize.Authorize(ctx, room, player.ID()); err != nil {
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
