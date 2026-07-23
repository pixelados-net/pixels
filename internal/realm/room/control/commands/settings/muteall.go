package settings

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	"github.com/niflaot/pixels/internal/realm/room/control/commands/resolve"
	muteallchanged "github.com/niflaot/pixels/internal/realm/room/control/events/muteallchanged"
	roomsettings "github.com/niflaot/pixels/internal/realm/room/control/settings"
	roommodel "github.com/niflaot/pixels/internal/realm/room/record/model"
	roomservice "github.com/niflaot/pixels/internal/realm/room/record/service"
	"github.com/niflaot/pixels/internal/realm/room/runtime/broadcast"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	netconn "github.com/niflaot/pixels/networking/connection"
	outstate "github.com/niflaot/pixels/networking/outbound/room/mute/state"
	"github.com/niflaot/pixels/pkg/bus"
)

const (
	// MuteAllName identifies the room mute-all toggle command.
	MuteAllName command.Name = "room.mute_all.toggle"
)

// MuteAllCommand toggles the actor's current room mute-all state.
type MuteAllCommand struct {
	// Handler stores the source connection.
	Handler netconn.Context
}

// MuteAllHandler handles room mute-all toggles.
type MuteAllHandler struct {
	// Players stores live players.
	Players *playerlive.Registry
	// Bindings stores connection bindings.
	Bindings *binding.Registry
	// Rooms reads room metadata.
	Rooms MuteRoomFinder
	// Authorize resolves settings capability.
	Authorize *roomsettings.Authorizer
	// Runtime stores active rooms.
	Runtime *roomlive.Registry
	// Connections stores active connections.
	Connections *netconn.Registry
	// Events publishes mute-all transitions.
	Events bus.Publisher
}

// MuteRoomFinder reads room metadata.
type MuteRoomFinder interface {
	// FindByID finds one room.
	FindByID(context.Context, int64) (roommodel.Room, bool, error)
}

// CommandName returns the stable command name.
func (MuteAllCommand) CommandName() command.Name { return MuteAllName }

// Handle toggles active room mute-all state.
func (handler MuteAllHandler) Handle(ctx context.Context, envelope command.Envelope[MuteAllCommand]) error {
	player, roomID, err := control.Actor(envelope.Command.Handler, handler.Bindings, handler.Players)
	if err != nil {
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
	if handler.Runtime == nil {
		return nil
	}
	active, found := handler.Runtime.Find(roomID)
	if !found {
		return nil
	}
	muted := !active.MuteAll()
	active.SetMuteAll(muted)
	packet, err := outstate.Encode(muted)
	if err != nil {
		return err
	}
	_ = broadcast.RoomPacket(ctx, handler.Connections, active, packet, 0)
	if handler.Events == nil {
		return nil
	}

	return handler.Events.Publish(ctx, bus.Event{Name: muteallchanged.Name, Payload: muteallchanged.Payload{RoomID: roomID, ActorID: player.ID(), Muted: muted}})
}
