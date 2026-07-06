// Package live contains runtime player state.
package live

import "errors"

var (
	// ErrInvalidPlayer reports an incomplete live player.
	ErrInvalidPlayer = errors.New("invalid live player")

	// ErrInvalidPeer reports an incomplete session peer.
	ErrInvalidPeer = errors.New("invalid session peer")

	// ErrPlayerExists reports a duplicate live player.
	ErrPlayerExists = errors.New("live player exists")

	// ErrInvalidRoomPresence reports malformed room presence data.
	ErrInvalidRoomPresence = errors.New("invalid room presence")
)
