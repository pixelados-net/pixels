package live

import (
	"errors"
	"time"

	"github.com/niflaot/pixels/networking/connection"
)

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

// SessionPeer binds a live player to a connection identity.
type SessionPeer struct {
	// connectionID identifies the live connection.
	connectionID connection.ID

	// connectionKind identifies the connection family.
	connectionKind connection.Kind

	// authenticatedAt stores when authentication completed.
	authenticatedAt time.Time
}

// NewSessionPeer creates a player session peer.
func NewSessionPeer(id connection.ID, kind connection.Kind, authenticatedAt time.Time) (SessionPeer, error) {
	peer := SessionPeer{connectionID: id, connectionKind: kind, authenticatedAt: authenticatedAt}
	if !peer.Valid() {
		return SessionPeer{}, ErrInvalidPeer
	}

	return peer, nil
}

// Valid reports whether the peer can bind a live player.
func (peer SessionPeer) Valid() bool {
	return peer.connectionID != "" && peer.connectionKind != "" && !peer.authenticatedAt.IsZero()
}

// ConnectionID returns the connection identifier.
func (peer SessionPeer) ConnectionID() connection.ID {
	return peer.connectionID
}

// ConnectionKind returns the connection kind.
func (peer SessionPeer) ConnectionKind() connection.Kind {
	return peer.connectionKind
}

// AuthenticatedAt returns when authentication completed.
func (peer SessionPeer) AuthenticatedAt() time.Time {
	return peer.authenticatedAt
}
