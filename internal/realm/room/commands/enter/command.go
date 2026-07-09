// Package enter joins a player into an active room.
package enter

import (
	"context"
	"errors"

	"github.com/niflaot/pixels/internal/command"
	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	playerservice "github.com/niflaot/pixels/internal/realm/player/service"
	roomsession "github.com/niflaot/pixels/internal/realm/room/commands/session"
	"github.com/niflaot/pixels/internal/realm/room/layout"
	roomlive "github.com/niflaot/pixels/internal/realm/room/live"
	roommodel "github.com/niflaot/pixels/internal/realm/room/model"
	roomservice "github.com/niflaot/pixels/internal/realm/room/service"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	netconn "github.com/niflaot/pixels/networking/connection"
	outentered "github.com/niflaot/pixels/networking/outbound/room/entered"
	outentryerror "github.com/niflaot/pixels/networking/outbound/room/entryerror"
	"github.com/niflaot/pixels/pkg/bus"
)

const (
	// Name identifies the room enter command.
	Name command.Name = "room.enter"
	// ErrorRoomFull is the protocol room-full error code.
	ErrorRoomFull int32 = 1
)

// Command joins a room.
type Command struct {
	// Handler stores the source connection handler.
	Handler netconn.Context
	// RoomID identifies the room to join.
	RoomID int64
	// Password stores the optional room entry password.
	Password string
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
}

// CommandName returns the stable command name.
func (Command) CommandName() command.Name {
	return Name
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
	active, err := handler.join(ctx, player, envelope.Command.Handler, room, roomLayout)
	if err != nil {
		return handler.sendEntryError(ctx, envelope.Command.Handler, err)
	}
	if err := player.EnterRoom(room.ID); err != nil {
		return err
	}

	if err := handler.sendEntered(ctx, envelope.Command.Handler, room, roomLayout, active); err != nil {
		return err
	}

	return handler.broadcastJoined(ctx, active, player.ID())
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

	roomLayout, found, err := handler.Layouts.FindByName(ctx, room.ModelName)
	if err != nil {
		return roommodel.Room{}, layout.Layout{}, err
	}
	if !found {
		return roommodel.Room{}, layout.Layout{}, roomservice.ErrLayoutNotAvailable
	}

	return room, roomLayout, nil
}

// sendEntered sends the initial room entry packets.
func (handler Handler) sendEntered(ctx context.Context, connection netconn.Context, room roommodel.Room, roomLayout layout.Layout, active *roomlive.Room) error {
	packet, err := outentered.Encode()
	if err != nil {
		return err
	}
	if err := connection.Send(ctx, packet); err != nil {
		return err
	}

	if err := SendModel(ctx, connection, room, roomLayout); err != nil {
		return err
	}

	if err := handler.sendFloorItems(ctx, connection, room, active); err != nil {
		return err
	}
	if err := handler.sendHeightMap(ctx, connection, active); err != nil {
		return err
	}

	return handler.sendRoomState(ctx, connection, active, 0)
}

// sendEntryError sends a room entry error when possible.
func (handler Handler) sendEntryError(ctx context.Context, connection netconn.Context, err error) error {
	if !errors.Is(err, roomlive.ErrRoomFull) {
		return err
	}

	packet, encodeErr := outentryerror.Encode(ErrorRoomFull)
	if encodeErr != nil {
		return encodeErr
	}

	return connection.Send(ctx, packet)
}
