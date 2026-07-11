// Package request sends editable room settings.
package request

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	"github.com/niflaot/pixels/internal/realm/room/commands/control"
	roommodel "github.com/niflaot/pixels/internal/realm/room/model"
	roomservice "github.com/niflaot/pixels/internal/realm/room/service"
	roomsettings "github.com/niflaot/pixels/internal/realm/room/settings"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	netconn "github.com/niflaot/pixels/networking/connection"
	outcurrent "github.com/niflaot/pixels/networking/outbound/room/settings/current"
)

const (
	// Name identifies the room settings request command.
	Name command.Name = "room.settings.request"
)

// Command requests one room settings snapshot.
type Command struct {
	// Handler stores the source connection.
	Handler netconn.Context
	// RoomID identifies the requested room.
	RoomID int64
}

// Handler handles room settings requests.
type Handler struct {
	// Players stores live players.
	Players *playerlive.Registry
	// Bindings stores connection bindings.
	Bindings *binding.Registry
	// Rooms reads room persistence.
	Rooms RoomReader
	// Authorize resolves settings capability.
	Authorize *roomsettings.Authorizer
}

// RoomReader reads settings room metadata and tags.
type RoomReader interface {
	// FindByID finds one room.
	FindByID(context.Context, int64) (roommodel.Room, bool, error)
	// ListTags lists room tags.
	ListTags(context.Context, int64) ([]roommodel.Tag, error)
}

// CommandName returns the stable command name.
func (Command) CommandName() command.Name { return Name }

// Handle sends editable room settings.
func (handler Handler) Handle(ctx context.Context, envelope command.Envelope[Command]) error {
	player, roomID, err := control.Actor(envelope.Command.Handler, handler.Bindings, handler.Players)
	if err != nil {
		return err
	}
	if err = control.MatchRoom(roomID, envelope.Command.RoomID); err != nil {
		return err
	}
	room, found, err := handler.Rooms.FindByID(ctx, roomID)
	if err != nil {
		return err
	}
	if !found {
		return roomservice.ErrRoomNotFound
	}
	if err = handler.Authorize.Authorize(ctx, room, player.ID()); err != nil {
		return err
	}
	tags, err := handler.Rooms.ListTags(ctx, roomID)
	if err != nil {
		return err
	}
	packet, err := outcurrent.Encode(settingsParams(room, tags))
	if err != nil {
		return err
	}

	return envelope.Command.Handler.Send(ctx, packet)
}

// settingsParams projects persistent room settings to protocol fields.
func settingsParams(room roommodel.Room, tags []roommodel.Tag) outcurrent.Params {
	values := make([]string, 0, len(tags))
	for _, tag := range tags {
		values = append(values, tag.Value)
	}
	categoryID := int32(0)
	if room.CategoryID != nil {
		categoryID = int32(*room.CategoryID)
	}

	return outcurrent.Params{RoomID: int32(room.ID), Name: room.Name, Description: room.Description,
		DoorMode: int32(room.DoorMode), CategoryID: categoryID, MaxUsers: int32(room.MaxUsers), MaxUsersLimit: roomservice.MaxRoomUsers,
		Tags: values, TradeMode: int32(room.TradeMode), AllowPets: room.AllowPets, AllowPetsEat: room.AllowPetsEat,
		AllowWalkthrough: room.AllowWalkthrough, HideWalls: room.HideWalls, WallThickness: int32(room.WallThickness),
		FloorThickness: int32(room.FloorThickness), ChatMode: int32(room.ChatMode), ChatWeight: int32(room.ChatWeight),
		ChatSpeed: int32(room.ChatSpeed), ChatDistance: int32(room.ChatDistance), ChatProtection: int32(room.ChatProtection),
		ModerationMute: int32(room.ModerationMute), ModerationKick: int32(room.ModerationKick), ModerationBan: int32(room.ModerationBan)}
}
