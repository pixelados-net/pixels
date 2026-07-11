// Package info sends navigator room information.
package info

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	navsession "github.com/niflaot/pixels/internal/realm/navigator/commands/session"
	navprojection "github.com/niflaot/pixels/internal/realm/navigator/projection"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	roomlive "github.com/niflaot/pixels/internal/realm/room/live"
	roommodel "github.com/niflaot/pixels/internal/realm/room/model"
	roommoderation "github.com/niflaot/pixels/internal/realm/room/moderation"
	moderationmodel "github.com/niflaot/pixels/internal/realm/room/moderation/model"
	roomservice "github.com/niflaot/pixels/internal/realm/room/service"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	netconn "github.com/niflaot/pixels/networking/connection"
	outnosuch "github.com/niflaot/pixels/networking/outbound/navigator/nosuchflat"
	outinfo "github.com/niflaot/pixels/networking/outbound/navigator/roominfo"
)

const (
	// Name identifies the navigator room info command.
	Name command.Name = "navigator.room_info"
)

// Command sends room information.
type Command struct {
	// Handler stores the source connection handler.
	Handler netconn.Context
	// RoomID identifies the requested room.
	RoomID int64
	// EnterRoom reports whether the client intends to enter.
	EnterRoom bool
	// ForwardRoom reports whether this is a forward flow.
	ForwardRoom bool
}

// Handler handles room information commands.
type Handler struct {
	// Players stores live player state.
	Players *playerlive.Registry
	// Bindings stores player connection bindings.
	Bindings *binding.Registry
	// Rooms reads room persistence.
	Rooms roomservice.Manager
	// Runtime reads active room occupancy.
	Runtime *roomlive.Registry
	// Moderation resolves viewer room moderation capability.
	Moderation roommoderation.Manager
}

// CommandName returns the stable command name.
func (Command) CommandName() command.Name {
	return Name
}

// Handle handles a room information command.
func (handler Handler) Handle(ctx context.Context, envelope command.Envelope[Command]) error {
	player, _, err := navsession.Player(envelope.Command.Handler, handler.Bindings, handler.Players)
	if err != nil {
		return err
	}

	room, found, err := handler.Rooms.FindByID(ctx, envelope.Command.RoomID)
	if err != nil {
		return err
	}
	if !found {
		return sendNoSuchRoom(ctx, envelope.Command.Handler)
	}

	tags, err := handler.roomTags(ctx, room.ID)
	if err != nil {
		return err
	}

	return handler.sendRoomInfo(ctx, envelope.Command, room, player.ID(), tags)
}

// sendRoomInfo sends one navigator room info packet.
func (handler Handler) sendRoomInfo(ctx context.Context, input Command, room roommodel.Room, viewerID int64, tags []string) error {
	canMute := false
	if handler.Moderation != nil {
		var err error
		canMute, err = handler.Moderation.CanModerate(ctx, room, viewerID, moderationmodel.ActionMute)
		if err != nil {
			return err
		}
	}
	packet, err := outinfo.Encode(outinfo.Params{
		RoomEnter:      input.EnterRoom,
		Room:           navprojection.RoomCard(room, handler.userCount(room.ID), 0, tags),
		RoomForward:    input.ForwardRoom,
		StaffPick:      room.StaffPicked,
		Moderation:     moderation(room),
		CanMute:        canMute,
		AllInRoomMuted: handler.allInRoomMuted(room.ID),
		Chat:           chat(room),
	})
	if err != nil {
		return err
	}

	return input.Handler.Send(ctx, packet)
}

// allInRoomMuted reads active room mute-all state.
func (handler Handler) allInRoomMuted(roomID int64) bool {
	if handler.Runtime == nil {
		return false
	}
	active, found := handler.Runtime.Find(roomID)
	if !found {
		return false
	}

	return active.MuteAll()
}

// sendNoSuchRoom sends a missing room response.
func sendNoSuchRoom(ctx context.Context, handler netconn.Context) error {
	packet, err := outnosuch.Encode(0)
	if err != nil {
		return err
	}

	return handler.Send(ctx, packet)
}
