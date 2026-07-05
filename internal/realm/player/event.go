// Package player contains player realm wiring.
package player

import (
	"github.com/niflaot/pixels/networking/connection"
	"github.com/niflaot/pixels/pkg/bus"
)

const (
	// EventAuthenticating reports an authentication attempt for a player.
	EventAuthenticating bus.Name = "player.authenticating"

	// EventAuthenticated reports a successfully authenticated player.
	EventAuthenticated bus.Name = "player.authenticated"

	// EventAuthenticationFailed reports a rejected player authentication.
	EventAuthenticationFailed bus.Name = "player.authentication_failed"

	// EventConnected reports a live player connection became ready.
	EventConnected bus.Name = "player.connected"

	// EventDisconnected reports a live player connection was disposed.
	EventDisconnected bus.Name = "player.disconnected"

	// EventProfileLoaded reports a player profile loaded into runtime state.
	EventProfileLoaded bus.Name = "player.profile_loaded"
)

// AuthenticationEvent describes a player authentication lifecycle event.
type AuthenticationEvent struct {
	// PlayerID identifies the player when known.
	PlayerID int64

	// ConnectionID identifies the connection.
	ConnectionID connection.ID

	// ConnectionKind identifies the connection family.
	ConnectionKind connection.Kind

	// Reason stores a failure reason when available.
	Reason string
}
