// Package session resolves catalog command session context.
package session

import (
	"errors"
	"time"

	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	netconn "github.com/niflaot/pixels/networking/connection"
)

var (
	// ErrBindingNotFound reports a missing live connection binding.
	ErrBindingNotFound = errors.New("catalog session binding not found")
	// ErrPlayerNotFound reports a missing live player.
	ErrPlayerNotFound = errors.New("catalog live player not found")
)

// Player resolves the player bound to a connection handler.
func Player(handler netconn.Context, bindings *binding.Registry, players *playerlive.Registry) (*playerlive.Player, error) {
	if bindings == nil || players == nil {
		return nil, ErrBindingNotFound
	}
	sessionBinding, found := bindings.FindByConnection(binding.ConnectionKey{Kind: handler.ConnectionKind, ID: handler.ConnectionID})
	if !found {
		return nil, ErrBindingNotFound
	}
	player, found := players.Find(sessionBinding.PlayerID)
	if !found {
		return nil, ErrPlayerNotFound
	}

	return player, nil
}

// HasClub reports whether a live player has an active club entitlement.
func HasClub(player *playerlive.Player) bool {
	return player != nil && player.Snapshot().HasClubAt(time.Now())
}
