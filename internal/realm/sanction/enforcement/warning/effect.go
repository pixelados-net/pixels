// Package warning delivers online and pending sanction warnings.
package warning

import (
	"context"

	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	sanctionrecord "github.com/niflaot/pixels/internal/realm/sanction/record"
	netconn "github.com/niflaot/pixels/networking/connection"
	outalert "github.com/niflaot/pixels/networking/outbound/session/alert"
)

// Warn delivers warnings online or persists them for login.
type Warn struct {
	// store persists offline alerts.
	store sanctionrecord.Store
	// players locates authenticated recipients.
	players *playerlive.Registry
	// connections delivers alert packets.
	connections *netconn.Registry
}

// NewWarn creates a warning behavior.
func NewWarn(store sanctionrecord.Store, players *playerlive.Registry, connections *netconn.Registry) *Warn {
	return &Warn{store: store, players: players, connections: connections}
}

// Kind identifies warning behavior.
func (*Warn) Kind() sanctionrecord.Kind { return sanctionrecord.KindWarn }

// Apply delivers or queues one warning.
func (effect *Warn) Apply(ctx context.Context, punishment sanctionrecord.Punishment) error {
	packet, err := outalert.Encode(punishment.Reason)
	if err != nil {
		return err
	}
	player, found := effect.players.Find(punishment.ReceiverPlayerID)
	if found {
		peer := player.Peer()
		connection, connected := effect.connections.Get(peer.ConnectionKind(), peer.ConnectionID())
		if connected {
			return connection.Send(ctx, packet)
		}
	}
	id := punishment.ID
	return effect.store.QueueAlert(ctx, sanctionrecord.Alert{PlayerID: punishment.ReceiverPlayerID, PunishmentID: &id, Message: punishment.Reason})
}

// Revoke has no effect on an already delivered warning.
func (*Warn) Revoke(context.Context, sanctionrecord.Punishment) error { return nil }
