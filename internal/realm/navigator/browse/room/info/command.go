// Package info sends navigator room information.
package info

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	grouprecord "github.com/niflaot/pixels/internal/realm/group/record"
	navprojection "github.com/niflaot/pixels/internal/realm/navigator/browse/card"
	navsession "github.com/niflaot/pixels/internal/realm/navigator/session"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	roommoderation "github.com/niflaot/pixels/internal/realm/room/control/moderation"
	moderationmodel "github.com/niflaot/pixels/internal/realm/room/control/moderation/model"
	roommodel "github.com/niflaot/pixels/internal/realm/room/record/model"
	roomservice "github.com/niflaot/pixels/internal/realm/room/record/service"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	netconn "github.com/niflaot/pixels/networking/connection"
	outnosuch "github.com/niflaot/pixels/networking/outbound/navigator/browse/nosuchflat"
	"github.com/niflaot/pixels/networking/outbound/navigator/browse/roomcard"
	outinfo "github.com/niflaot/pixels/networking/outbound/navigator/browse/roominfo"
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
	// Groups resolves viewer-specific room group data.
	Groups RoomGroups
}

// RoomGroups resolves one room's group without loading its roster.
type RoomGroups interface {
	// RoomGroupInfo returns group metadata and viewer membership.
	RoomGroupInfo(context.Context, int64, int64) (grouprecord.Group, bool, bool, error)
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
	card := navprojection.RoomCard(room, handler.userCount(room.ID), 0, tags)
	isGroupMember := false
	if handler.Groups != nil {
		group, member, found, groupErr := handler.Groups.RoomGroupInfo(ctx, room.ID, viewerID)
		if groupErr != nil {
			return groupErr
		}
		if found {
			card.Group = &roomcard.Group{ID: int32(group.ID), Name: group.Name, Badge: group.BadgeCode}
			isGroupMember = member
		}
	}
	packet, err := outinfo.Encode(outinfo.Params{
		RoomEnter:      input.EnterRoom,
		Room:           card,
		IsGroupMember:  isGroupMember,
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
