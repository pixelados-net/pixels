// Package broadcast projects committed moderation actions into room runtime.
package broadcast

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	"github.com/niflaot/pixels/internal/realm/room/broadcast"
	leavecmd "github.com/niflaot/pixels/internal/realm/room/commands/leave"
	roomlive "github.com/niflaot/pixels/internal/realm/room/live"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	outentryerror "github.com/niflaot/pixels/networking/outbound/room/entryerror"
	outmuted "github.com/niflaot/pixels/networking/outbound/room/moderation/muted"
	outunbanned "github.com/niflaot/pixels/networking/outbound/room/moderation/unbanned"
	outerror "github.com/niflaot/pixels/networking/outbound/session/error"
	"github.com/niflaot/pixels/pkg/bus"
)

const (
	// kickedErrorCode identifies Nitro's kicked-out-of-room generic error.
	kickedErrorCode int32 = 4008
	// bannedEntryErrorCode identifies a room ban entry error.
	bannedEntryErrorCode int32 = 4
)

// Broadcaster projects committed moderation actions.
type Broadcaster struct {
	// runtime stores active rooms.
	runtime *roomlive.Registry
	// connections stores active connections.
	connections *netconn.Registry
	// leave removes targets through standard room lifecycle.
	leave leavecmd.Handler
}

// New creates a room moderation broadcaster.
func New(players *playerlive.Registry, bindings *binding.Registry, runtime *roomlive.Registry, connections *netconn.Registry, events bus.Publisher) *Broadcaster {
	return &Broadcaster{
		runtime: runtime, connections: connections,
		leave: leavecmd.Handler{Players: players, Bindings: bindings, Runtime: runtime, Connections: connections, Events: events},
	}
}

// Kick notifies and removes one active room occupant.
func (broadcaster *Broadcaster) Kick(ctx context.Context, roomID int64, playerID int64) error {
	packet, err := outerror.Encode(kickedErrorCode)
	if err != nil {
		return err
	}

	return broadcaster.remove(ctx, roomID, playerID, packet)
}

// Ban notifies and removes one active room occupant.
func (broadcaster *Broadcaster) Ban(ctx context.Context, roomID int64, playerID int64) error {
	packet, err := outentryerror.Encode(bannedEntryErrorCode)
	if err != nil {
		return err
	}

	return broadcaster.remove(ctx, roomID, playerID, packet)
}

// Mute sends the remaining mute duration to one active occupant.
func (broadcaster *Broadcaster) Mute(ctx context.Context, roomID int64, playerID int64, seconds int64) error {
	packet, err := outmuted.Encode(int32(seconds))
	if err != nil {
		return err
	}

	return broadcaster.sendTarget(ctx, roomID, playerID, packet)
}

// Unban broadcasts removal from a room ban list.
func (broadcaster *Broadcaster) Unban(ctx context.Context, roomID int64, playerID int64) error {
	active, found := broadcaster.runtime.Find(roomID)
	if !found {
		return nil
	}
	packet, err := outunbanned.Encode(int32(roomID), int32(playerID))
	if err != nil {
		return err
	}

	return broadcast.RoomPacket(ctx, broadcaster.connections, active, packet, 0)
}

// remove sends one protocol notice and leaves through standard room lifecycle.
func (broadcaster *Broadcaster) remove(ctx context.Context, roomID int64, playerID int64, packet codec.Packet) error {
	active, found := broadcaster.runtime.FindByPlayer(playerID)
	if !found || active.ID() != roomID {
		return nil
	}
	if err := broadcaster.sendTarget(ctx, roomID, playerID, packet); err != nil {
		return err
	}

	return broadcaster.leave.Handle(ctx, command.Envelope[leavecmd.Command]{Command: leavecmd.Command{PlayerID: playerID}})
}

// sendTarget sends one packet to a target only while present in the selected room.
func (broadcaster *Broadcaster) sendTarget(ctx context.Context, roomID int64, playerID int64, packet codec.Packet) error {
	if broadcaster.connections == nil {
		return nil
	}
	active, found := broadcaster.runtime.Find(roomID)
	if !found {
		return nil
	}
	for _, occupant := range active.Occupants() {
		if occupant.PlayerID != playerID {
			continue
		}
		connection, found := broadcaster.connections.Get(occupant.ConnectionKind, occupant.ConnectionID)
		if !found {
			return nil
		}

		return connection.Send(ctx, packet)
	}

	return nil
}
