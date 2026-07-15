// Package session enforces ban and kick against active connections.
package session

import (
	"context"

	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	sanctionrecord "github.com/niflaot/pixels/internal/realm/sanction/record"
	netconn "github.com/niflaot/pixels/networking/connection"
)

// SessionDisconnect executes ban or kick against an active session.
type SessionDisconnect struct {
	// kind identifies ban or kick.
	kind sanctionrecord.Kind
	// players locates authenticated peers.
	players *playerlive.Registry
	// connections owns connection disposal.
	connections *netconn.Registry
}

// NewBan creates a ban behavior.
func NewBan(players *playerlive.Registry, connections *netconn.Registry) *Ban {
	return &Ban{SessionDisconnect: SessionDisconnect{kind: sanctionrecord.KindBan, players: players, connections: connections}}
}

// NewKick creates a kick behavior.
func NewKick(players *playerlive.Registry, connections *netconn.Registry) *Kick {
	return &Kick{SessionDisconnect: SessionDisconnect{kind: sanctionrecord.KindKick, players: players, connections: connections}}
}

// Ban is the registered ban session behavior.
type Ban struct{ SessionDisconnect }

// Kick is the registered kick session behavior.
type Kick struct{ SessionDisconnect }

// Kind identifies the session effect.
func (effect *SessionDisconnect) Kind() sanctionrecord.Kind { return effect.kind }

// Apply disconnects an active target after the record commits.
func (effect *SessionDisconnect) Apply(ctx context.Context, punishment sanctionrecord.Punishment) error {
	player, found := effect.players.Find(punishment.ReceiverPlayerID)
	if !found {
		return nil
	}
	peer := player.Peer()
	code := netconn.DisconnectKicked
	if effect.kind == sanctionrecord.KindBan {
		code = netconn.DisconnectBanned
	}
	return effect.connections.Disconnect(ctx, peer.ConnectionKind(), peer.ConnectionID(), netconn.Reason{Code: code, Message: punishment.Reason})
}

// Revoke has no immediate connection effect.
func (*SessionDisconnect) Revoke(context.Context, sanctionrecord.Punishment) error { return nil }
