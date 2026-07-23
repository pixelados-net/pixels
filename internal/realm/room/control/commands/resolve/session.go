// Package control resolves the actor and room for room control commands.
package control

import (
	"errors"

	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	roomsession "github.com/niflaot/pixels/internal/realm/room/runtime/commands/session"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	netconn "github.com/niflaot/pixels/networking/connection"
)

var (
	// ErrPlayerNotInRoom reports an actor without active room presence.
	ErrPlayerNotInRoom = errors.New("room control player not in room")
	// ErrRoomMismatch reports a packet targeting another room.
	ErrRoomMismatch = errors.New("room control target does not match current room")
	// ErrTargetNotInRoom reports a protocol target outside the actor's room.
	ErrTargetNotInRoom = errors.New("room control target is not in room")
)

// Actor resolves an authenticated player and current room.
func Actor(handler netconn.Context, bindings *binding.Registry, players *playerlive.Registry) (*playerlive.Player, int64, error) {
	player, err := roomsession.Player(handler, bindings, players)
	if err != nil {
		return nil, 0, err
	}
	roomID, found := player.CurrentRoom()
	if !found {
		return nil, 0, ErrPlayerNotInRoom
	}

	return player, roomID, nil
}

// MatchRoom verifies a packet room id against current presence.
func MatchRoom(currentRoomID int64, packetRoomID int64) error {
	if packetRoomID <= 0 || currentRoomID != packetRoomID {
		return ErrRoomMismatch
	}

	return nil
}

// TargetInRoom verifies an active target belongs to the selected room.
func TargetInRoom(runtime *roomlive.Registry, roomID int64, playerID int64) error {
	if runtime == nil || playerID <= 0 {
		return ErrTargetNotInRoom
	}
	active, found := runtime.FindByPlayer(playerID)
	if !found || active.ID() != roomID {
		return ErrTargetNotInRoom
	}

	return nil
}
