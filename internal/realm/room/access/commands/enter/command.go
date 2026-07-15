// Package enter joins a player into an active room.
package enter

import (
	"context"
	"errors"

	"github.com/niflaot/pixels/internal/command"
	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	playerservice "github.com/niflaot/pixels/internal/realm/player/service"
	roomentry "github.com/niflaot/pixels/internal/realm/room/access/entry"
	roomentered "github.com/niflaot/pixels/internal/realm/room/access/events/entered"
	roommoderation "github.com/niflaot/pixels/internal/realm/room/control/moderation"
	roomrights "github.com/niflaot/pixels/internal/realm/room/control/rights"
	roomvotes "github.com/niflaot/pixels/internal/realm/room/control/votes"
	roommodel "github.com/niflaot/pixels/internal/realm/room/record/model"
	roomservice "github.com/niflaot/pixels/internal/realm/room/record/service"
	roomsession "github.com/niflaot/pixels/internal/realm/room/runtime/commands/session"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/internal/realm/room/world/layout"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	netconn "github.com/niflaot/pixels/networking/connection"
	outentryerror "github.com/niflaot/pixels/networking/outbound/room/entryerror"
	outalert "github.com/niflaot/pixels/networking/outbound/session/alert"
	outdesktop "github.com/niflaot/pixels/networking/outbound/session/desktop"
	outerror "github.com/niflaot/pixels/networking/outbound/session/error"
	"github.com/niflaot/pixels/pkg/bus"
	"go.uber.org/zap/zapcore"
)

const (
	// Name identifies the room enter command.
	Name command.Name = "room.enter"
	// ErrorRoomFull is the protocol room-full error code.
	ErrorRoomFull int32 = 1
	// ErrorAccessDenied is the protocol closed-room error code.
	ErrorAccessDenied int32 = 2
	// ErrorQueue is the protocol room-queue error code.
	ErrorQueue int32 = 3
	// ErrorBanned is the protocol room-ban error code.
	ErrorBanned int32 = 4
	// ErrorWrongPassword is the generic wrong-password error code.
	ErrorWrongPassword int32 = -100002
)

// Command joins a room.
type Command struct {
	// Handler stores the source connection handler.
	Handler netconn.Context
	// RoomID identifies the room to join.
	RoomID int64
	// Password stores the optional room entry password.
	Password string
	// Trusted marks a direct server-controlled entry.
	Trusted bool
}

// Handler handles room entry commands.
type Handler struct {
	// Players stores live player state.
	Players *playerlive.Registry
	// Bindings stores player connection bindings.
	Bindings *binding.Registry
	// Rooms reads room persistence.
	Rooms roomservice.Manager
	// Layouts reads room layouts.
	Layouts layout.Manager
	// Furniture reads placed and inventory furniture records.
	Furniture furnitureservice.Manager
	// PlayerDirectory resolves durable player identities for furniture owners not currently online.
	PlayerDirectory playerservice.Finder
	// Runtime stores active rooms.
	Runtime *roomlive.Registry
	// Connections stores active network connections.
	Connections *netconn.Registry
	// Events publishes room lifecycle events.
	Events bus.Publisher
	// Entry decides closed-room access.
	Entry *roomentry.Service
	// Rights manages persistent room build rights.
	Rights roomrights.Manager
	// Moderation reads current room mute projections during activation.
	Moderation roommoderation.Reader
	// Votes reads room score and player eligibility.
	Votes roomvotes.Reader
	// Control projects global room capabilities into Nitro controller state.
	Control ControlPolicy
}

// CommandName returns the stable command name.
func (Command) CommandName() command.Name {
	return Name
}

// MarshalLogObject writes command fields without exposing plaintext passwords.
func (command Command) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	encoder.AddInt64("room_id", command.RoomID)
	encoder.AddBool("password_provided", command.Password != "")
	encoder.AddBool("trusted", command.Trusted)

	return nil
}

// Handle handles a room enter command.
func (handler Handler) Handle(ctx context.Context, envelope command.Envelope[Command]) error {
	player, err := roomsession.Player(envelope.Command.Handler, handler.Bindings, handler.Players)
	if err != nil {
		return err
	}

	room, roomLayout, err := handler.loadRoom(ctx, envelope.Command.RoomID)
	if err != nil {
		return err
	}
	result, err := handler.authorize(ctx, roomentry.Request{
		Room: room, PlayerID: player.ID(), Password: envelope.Command.Password, Trusted: envelope.Command.Trusted,
	})
	if errors.Is(err, roomentry.ErrDoorbellRequired) {
		return handler.requestDoorbell(ctx, player, envelope.Command.Handler, room)
	}
	if err != nil {
		if result.Alert != "" {
			if sendErr := handler.sendAlert(ctx, envelope.Command.Handler, result.Alert); sendErr != nil {
				return sendErr
			}
		}

		return handler.sendEntryError(ctx, envelope.Command.Handler, err)
	}
	active, err := handler.join(ctx, player, envelope.Command.Handler, room, roomLayout)
	if err != nil {
		return handler.sendEntryError(ctx, envelope.Command.Handler, err)
	}
	if err := player.EnterRoom(room.ID); err != nil {
		return err
	}

	if err := handler.sendEntered(ctx, envelope.Command.Handler, room, roomLayout, active, player.ID()); err != nil {
		return err
	}
	if err := handler.publish(ctx, roomentered.Name, roomentered.Payload{PlayerID: player.ID(), RoomID: room.ID}); err != nil {
		return err
	}

	return handler.broadcastJoined(ctx, active, player.ID())
}

// authorize checks room entry policy when configured.
func (handler Handler) authorize(ctx context.Context, request roomentry.Request) (roomentry.Result, error) {
	if handler.Entry == nil {
		if request.Room.DoorMode == roommodel.DoorModeOpen {
			return roomentry.Result{}, nil
		}

		return roomentry.Result{}, roomentry.ErrAccessDenied
	}

	return handler.Entry.Authorize(ctx, request)
}

// sendAlert sends a localized entry protection message.
func (handler Handler) sendAlert(ctx context.Context, connection netconn.Context, message string) error {
	packet, err := outalert.Encode(message)
	if err != nil {
		return err
	}

	return connection.Send(ctx, packet)
}

// loadRoom loads room and layout data.
func (handler Handler) loadRoom(ctx context.Context, roomID int64) (roommodel.Room, layout.Layout, error) {
	room, found, err := handler.Rooms.FindByID(ctx, roomID)
	if err != nil {
		return roommodel.Room{}, layout.Layout{}, err
	}
	if !found {
		return roommodel.Room{}, layout.Layout{}, roomservice.ErrRoomNotFound
	}

	roomLayout, err := layout.ResolveForRoom(ctx, handler.Layouts, room.ID, room.ModelName)
	if err != nil {
		if errors.Is(err, layout.ErrLayoutNotFound) {
			return roommodel.Room{}, layout.Layout{}, roomservice.ErrLayoutNotAvailable
		}
		return roommodel.Room{}, layout.Layout{}, err
	}

	return room, roomLayout, nil
}

// sendEntryError sends a room entry error when possible.
func (handler Handler) sendEntryError(ctx context.Context, connection netconn.Context, err error) error {
	if errors.Is(err, roomentry.ErrEntryLocked) {
		packet, encodeErr := outdesktop.Encode()
		if encodeErr != nil {
			return encodeErr
		}

		return connection.Send(ctx, packet)
	}
	if errors.Is(err, roomentry.ErrWrongPassword) {
		packet, encodeErr := outerror.Encode(ErrorWrongPassword)
		if encodeErr != nil {
			return encodeErr
		}

		return connection.Send(ctx, packet)
	}
	code, found := entryErrorCode(err)
	if !found {
		return err
	}

	packet, encodeErr := outentryerror.Encode(code)
	if encodeErr != nil {
		return encodeErr
	}

	return connection.Send(ctx, packet)
}
