// Package broadcast projects currency events to live player connections.
package broadcast

import (
	"context"

	currencychanged "github.com/niflaot/pixels/internal/realm/inventory/currency/events/changed"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	outcredits "github.com/niflaot/pixels/networking/outbound/user/currency/credits"
	outnotification "github.com/niflaot/pixels/networking/outbound/user/currency/notification"
	"github.com/niflaot/pixels/pkg/bus"
)

const (
	// creditsType identifies the protocol credits sentinel.
	creditsType int32 = -1
)

// Broadcaster sends currency changes to the affected live player.
type Broadcaster struct {
	// players stores live player compositions.
	players *playerlive.Registry

	// connections stores active protocol connections.
	connections *netconn.Registry
}

// New creates a currency change broadcaster.
func New(players *playerlive.Registry, connections *netconn.Registry) *Broadcaster {
	return &Broadcaster{players: players, connections: connections}
}

// Handle projects one committed currency change.
func (broadcaster *Broadcaster) Handle(ctx context.Context, event bus.Event) error {
	payload, ok := event.Payload.(currencychanged.Payload)
	if !ok || payload.PlayerID <= 0 {
		return nil
	}

	connection, found := broadcaster.connection(payload.PlayerID)
	if !found {
		return nil
	}

	packet, err := changePacket(payload)
	if err != nil {
		return err
	}

	return connection.Send(ctx, packet)
}

// connection resolves one live player's active connection.
func (broadcaster *Broadcaster) connection(playerID int64) (netconn.Connection, bool) {
	player, found := broadcaster.players.Find(playerID)
	if !found {
		return nil, false
	}

	peer := player.Peer()

	return broadcaster.connections.Get(peer.ConnectionKind(), peer.ConnectionID())
}

// changePacket creates the protocol projection for one currency change.
func changePacket(payload currencychanged.Payload) (codec.Packet, error) {
	if payload.CurrencyType == creditsType {
		return outcredits.Encode(payload.Amount)
	}

	return outnotification.Encode(payload.Amount, payload.Delta, payload.CurrencyType)
}
