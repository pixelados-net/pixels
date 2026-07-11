package settings

import (
	"context"
	"errors"
	"time"

	"github.com/niflaot/pixels/internal/command"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	"github.com/niflaot/pixels/internal/realm/room/control/commands/resolve"
	settingsupdated "github.com/niflaot/pixels/internal/realm/room/control/events/settingsupdated"
	roomsettings "github.com/niflaot/pixels/internal/realm/room/control/settings"
	roommodel "github.com/niflaot/pixels/internal/realm/room/record/model"
	roomservice "github.com/niflaot/pixels/internal/realm/room/record/service"
	"github.com/niflaot/pixels/internal/realm/room/runtime/broadcast"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	netconn "github.com/niflaot/pixels/networking/connection"
	outchat "github.com/niflaot/pixels/networking/outbound/room/chatsettings/updated"
	outerror "github.com/niflaot/pixels/networking/outbound/room/settings/error"
	outsaved "github.com/niflaot/pixels/networking/outbound/room/settings/saved"
	outupdated "github.com/niflaot/pixels/networking/outbound/room/settings/updated"
	outthickness "github.com/niflaot/pixels/networking/outbound/room/thickness/updated"
	"github.com/niflaot/pixels/pkg/bus"
)

const (
	// SaveName identifies the room settings save command.
	SaveName command.Name = "room.settings.save"
)

// SaveCommand contains a complete client room settings mutation.
type SaveCommand struct {
	// Handler stores the source connection.
	Handler netconn.Context
	// RoomID identifies the room.
	RoomID int64
	// Name stores the visible room name.
	Name string
	// Description stores the room description.
	Description string
	// DoorMode stores room access mode.
	DoorMode int32
	// Password stores an optional replacement password.
	Password string
	// MaxUsers stores room capacity.
	MaxUsers int
	// CategoryID identifies the navigator category.
	CategoryID int64
	// Tags stores the complete room tag set.
	Tags []string
	// TradeMode stores trading mode.
	TradeMode int32
	// AllowPets reports whether pets may enter.
	AllowPets bool
	// AllowPetsEat reports whether pets may eat.
	AllowPetsEat bool
	// AllowWalkthrough reports whether units may overlap.
	AllowWalkthrough bool
	// HideWalls reports whether walls are hidden.
	HideWalls bool
	// WallThickness stores wall thickness.
	WallThickness int
	// FloorThickness stores floor thickness.
	FloorThickness int
	// ModerationMute stores mute policy.
	ModerationMute int32
	// ModerationKick stores kick policy.
	ModerationKick int32
	// ModerationBan stores ban policy.
	ModerationBan int32
	// ChatMode stores chat mode.
	ChatMode int16
	// ChatWeight stores bubble weight.
	ChatWeight int16
	// ChatSpeed stores scroll speed.
	ChatSpeed int16
	// ChatDistance stores full hearing range.
	ChatDistance int16
	// ChatProtection stores flood protection.
	ChatProtection int16
}

// SaveHandler handles room settings saves.
type SaveHandler struct {
	// Players stores live players.
	Players *playerlive.Registry
	// Bindings stores connection bindings.
	Bindings *binding.Registry
	// Rooms manages persistent room settings.
	Rooms SaveRoomManager
	// Authorize resolves settings capability.
	Authorize *roomsettings.Authorizer
	// Runtime stores active room state.
	Runtime *roomlive.Registry
	// Connections stores active connections.
	Connections *netconn.Registry
	// Events publishes committed settings events.
	Events bus.Publisher
}

// SaveRoomManager reads and updates room settings.
type SaveRoomManager interface {
	// FindByID finds one room.
	FindByID(context.Context, int64) (roommodel.Room, bool, error)
	// Update applies one optimistic room settings mutation.
	Update(context.Context, int64, int64, roomservice.UpdateParams) (roommodel.Room, error)
}

// CommandName returns the stable command name.
func (SaveCommand) CommandName() command.Name { return SaveName }

// Handle validates, persists, and broadcasts room settings.
func (handler SaveHandler) Handle(ctx context.Context, envelope command.Envelope[SaveCommand]) error {
	input := envelope.Command
	player, roomID, err := control.Actor(input.Handler, handler.Bindings, handler.Players)
	if err != nil {
		return err
	}
	if err = control.MatchRoom(roomID, input.RoomID); err != nil {
		return err
	}
	room, found, err := handler.Rooms.FindByID(ctx, roomID)
	if err != nil {
		return err
	}
	if !found {
		return handler.sendError(ctx, input.Handler, roomID, roomservice.ErrRoomNotFound)
	}
	if err = handler.Authorize.Authorize(ctx, room, player.ID()); err != nil {
		return handler.sendError(ctx, input.Handler, roomID, err)
	}
	allowReserved, err := handler.Authorize.CanManageAny(ctx, player.ID())
	if err != nil {
		return err
	}
	if moderationFieldsChanged(room, input) {
		if err = handler.Authorize.AuthorizePolicy(ctx, room, player.ID()); err != nil {
			return handler.sendError(ctx, input.Handler, roomID, err)
		}
	}
	if !clubFieldsAllowed(room, input, player.Snapshot().HasClubAt(time.Now()), allowReserved) {
		return roomsettings.ErrClubRequired
	}
	updated, err := handler.Rooms.Update(ctx, roomID, room.Version.Version, updateParams(input, allowReserved))
	if err != nil {
		return handler.sendError(ctx, input.Handler, roomID, err)
	}
	if handler.Runtime != nil {
		active, activeFound := handler.Runtime.Find(roomID)
		if activeFound {
			active.UpdateSettings(updated.CategoryID, updated.MaxUsers, updated.ChatDistance, updated.ChatProtection)
			if err = handler.broadcast(ctx, active, updated); err != nil {
				return err
			}
		}
	}
	packet, err := outsaved.Encode(int32(roomID))
	if err != nil {
		return err
	}
	if err = input.Handler.Send(ctx, packet); err != nil {
		return err
	}
	if handler.Events == nil {
		return nil
	}

	return handler.Events.Publish(ctx, bus.Event{Name: settingsupdated.Name, Payload: settingsupdated.Payload{RoomID: roomID, ActorID: player.ID(), Version: updated.Version.Version}})
}

// broadcast sends protocol-native settings refresh packets.
func (handler SaveHandler) broadcast(ctx context.Context, active *roomlive.Room, room roommodel.Room) error {
	updatedPacket, err := outupdated.Encode(int32(room.ID))
	if err != nil {
		return err
	}
	thicknessPacket, err := outthickness.Encode(room.HideWalls, int32(room.WallThickness), int32(room.FloorThickness))
	if err != nil {
		return err
	}
	chatPacket, err := outchat.Encode(int32(room.ChatMode), int32(room.ChatWeight), int32(room.ChatSpeed), int32(room.ChatDistance), int32(room.ChatProtection))
	if err != nil {
		return err
	}
	_ = broadcast.RoomPacket(ctx, handler.Connections, active, thicknessPacket, 0)
	_ = broadcast.RoomPacket(ctx, handler.Connections, active, chatPacket, 0)

	return broadcast.RoomPacket(ctx, handler.Connections, active, updatedPacket, 0)
}

// sendError maps domain errors to Nitro room settings error codes.
func (handler SaveHandler) sendError(ctx context.Context, connection netconn.Context, roomID int64, cause error) error {
	code := int32(outerror.CodeInvalidName)
	switch {
	case errors.Is(cause, roomservice.ErrRoomNotFound):
		code = outerror.CodeRoomNotFound
	case errors.Is(cause, roomsettings.ErrAccessDenied):
		code = outerror.CodeNotOwner
	case errors.Is(cause, roomservice.ErrInvalidDoorMode):
		code = outerror.CodeInvalidDoorMode
	case errors.Is(cause, roomservice.ErrInvalidMaxUsers):
		code = outerror.CodeInvalidUserLimit
	case errors.Is(cause, roomservice.ErrInvalidCategory):
		code = outerror.CodeInvalidCategory
	case errors.Is(cause, roomservice.ErrPasswordRequired):
		code = outerror.CodeInvalidPassword
	case errors.Is(cause, roomservice.ErrInvalidDescription):
		code = outerror.CodeInvalidDescription
	case errors.Is(cause, roomservice.ErrProhibitedName):
		code = outerror.CodeUnacceptableName
	case errors.Is(cause, roomservice.ErrProhibitedDescription):
		code = outerror.CodeUnacceptableDescription
	case errors.Is(cause, roomservice.ErrProhibitedTag):
		code = outerror.CodeInvalidTag
	case errors.Is(cause, roomservice.ErrReservedTag):
		code = outerror.CodeReservedTag
	case errors.Is(cause, roomservice.ErrInvalidTag):
		code = outerror.CodeInvalidTag
	}
	packet, err := outerror.Encode(int32(roomID), code, "")
	if err != nil {
		return err
	}

	return connection.Send(ctx, packet)
}
