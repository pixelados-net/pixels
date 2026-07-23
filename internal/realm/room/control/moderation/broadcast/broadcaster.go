// Package broadcast projects committed moderation actions into room runtime.
package broadcast

import (
	"context"
	"errors"
	"time"

	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	leavecmd "github.com/niflaot/pixels/internal/realm/room/access/commands/leave"
	"github.com/niflaot/pixels/internal/realm/room/runtime/broadcast"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	outentryerror "github.com/niflaot/pixels/networking/outbound/room/entryerror"
	outmuted "github.com/niflaot/pixels/networking/outbound/room/moderation/muted"
	outunbanned "github.com/niflaot/pixels/networking/outbound/room/moderation/unbanned"
	outalert "github.com/niflaot/pixels/networking/outbound/session/alert"
	"github.com/niflaot/pixels/pkg/bus"
	"github.com/niflaot/pixels/pkg/i18n"
)

const (
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
	// translations resolves end-user moderation notices.
	translations i18n.Translator
}

// New creates a room moderation broadcaster.
func New(players *playerlive.Registry, bindings *binding.Registry, runtime *roomlive.Registry, connections *netconn.Registry, events bus.Publisher, translations i18n.Translator) *Broadcaster {
	return &Broadcaster{
		runtime: runtime, connections: connections, translations: translations,
		leave: leavecmd.Handler{Players: players, Bindings: bindings, Runtime: runtime, Connections: connections, Events: events},
	}
}

// Kick notifies one occupant and walks it to the room door when reachable.
func (broadcaster *Broadcaster) Kick(ctx context.Context, roomID int64, playerID int64) error {
	packet, err := KickedNotice(broadcaster.translations)
	if err != nil {
		return err
	}
	active, found := broadcaster.runtime.FindByPlayer(playerID)
	if !found || active.ID() != roomID {
		return nil
	}
	walking, _ := active.ExitToDoor(playerID)
	if walking {
		return nil
	}

	return broadcaster.leave.ToDesktopThen(ctx, playerID, packet)
}

// KickedNotice creates a global localized notice that survives room teardown.
func KickedNotice(translations i18n.Translator) (codec.Packet, error) {
	message := "You were kicked out of the room."
	if translations != nil {
		message = translations.Default("room.moderation.kicked")
	}

	return outalert.Encode(message)
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
	active, found := broadcaster.runtime.Find(roomID)
	if found {
		endsAt := time.Time{}
		if seconds > 0 {
			endsAt = time.Now().Add(time.Duration(seconds) * time.Second)
		}
		active.SetMuted(playerID, endsAt)
	}
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
	target, targetFound := broadcaster.targetConnection(active, playerID)
	var noticeErr error
	if targetFound {
		noticeErr = target.Send(ctx, packet)
	}
	return errors.Join(noticeErr, broadcaster.leave.ToDesktop(ctx, playerID))
}

// sendTarget sends one packet to a target only while present in the selected room.
func (broadcaster *Broadcaster) sendTarget(ctx context.Context, roomID int64, playerID int64, packet codec.Packet) error {
	active, found := broadcaster.runtime.Find(roomID)
	if !found {
		return nil
	}
	connection, found := broadcaster.targetConnection(active, playerID)
	if !found {
		return nil
	}

	return connection.Send(ctx, packet)
}

// targetConnection resolves one occupant's connection before runtime removal.
func (broadcaster *Broadcaster) targetConnection(active *roomlive.Room, playerID int64) (netconn.Connection, bool) {
	if broadcaster.connections == nil || active == nil {
		return nil, false
	}
	for _, occupant := range active.Occupants() {
		if occupant.PlayerID != playerID {
			continue
		}
		connection, found := broadcaster.connections.Get(occupant.ConnectionKind, occupant.ConnectionID)

		return connection, found
	}

	return nil, false
}
