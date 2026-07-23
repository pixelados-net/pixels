// Package session resolves navigator command session context.
package session

import (
	"errors"

	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	netconn "github.com/niflaot/pixels/networking/connection"
)

var (
	// ErrBindingNotFound reports a missing live connection binding.
	ErrBindingNotFound = errors.New("navigator session binding not found")

	// ErrPlayerNotFound reports a missing live player.
	ErrPlayerNotFound = errors.New("navigator live player not found")
)

// Player resolves the player bound to a connection handler.
func Player(handler netconn.Context, bindings *binding.Registry, players *playerlive.Registry) (*playerlive.Player, binding.Binding, error) {
	if bindings == nil || players == nil {
		return nil, binding.Binding{}, ErrBindingNotFound
	}

	sessionBinding, found := bindings.FindByConnection(binding.ConnectionKey{Kind: handler.ConnectionKind, ID: handler.ConnectionID})
	if !found {
		return nil, binding.Binding{}, ErrBindingNotFound
	}

	player, found := players.Find(sessionBinding.PlayerID)
	if !found {
		return nil, binding.Binding{}, ErrPlayerNotFound
	}

	return player, sessionBinding, nil
}
