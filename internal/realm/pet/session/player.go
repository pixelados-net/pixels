// Package session resolves authenticated player state for pet commands.
package session

import (
	"errors"

	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	netconn "github.com/niflaot/pixels/networking/connection"
)

var (
	// ErrBindingNotFound reports a command without an authenticated binding.
	ErrBindingNotFound = errors.New("pet session binding not found")
	// ErrPlayerNotFound reports a command without live player state.
	ErrPlayerNotFound = errors.New("pet live player not found")
)

// Player resolves the live player bound to one packet connection.
func Player(handler netconn.Context, bindings *binding.Registry, players *playerlive.Registry) (*playerlive.Player, error) {
	if bindings == nil || players == nil {
		return nil, ErrBindingNotFound
	}
	value, found := bindings.FindByConnection(binding.ConnectionKey{Kind: handler.ConnectionKind, ID: handler.ConnectionID})
	if !found {
		return nil, ErrBindingNotFound
	}
	resolved, found := players.Find(value.PlayerID)
	if !found {
		return nil, ErrPlayerNotFound
	}
	return resolved, nil
}
