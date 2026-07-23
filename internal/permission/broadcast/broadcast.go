package broadcast

import (
	"context"
	"errors"

	permissionchanged "github.com/niflaot/pixels/internal/permission/events/changed"
	permissionservice "github.com/niflaot/pixels/internal/permission/service"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	netconn "github.com/niflaot/pixels/networking/connection"
	"github.com/niflaot/pixels/pkg/bus"
)

// Broadcaster sends permission changes to affected live players.
type Broadcaster struct {
	// projector sends protocol permission state.
	projector *Projector
	// permissions resolves group descendants.
	permissions permissionservice.Manager
	// players stores live player compositions.
	players *playerlive.Registry
	// connections stores active protocol connections.
	connections *netconn.Registry
}

// New creates a permission change broadcaster.
func New(projector *Projector, permissions permissionservice.Manager, players *playerlive.Registry, connections *netconn.Registry) *Broadcaster {
	return &Broadcaster{projector: projector, permissions: permissions, players: players, connections: connections}
}

// Handle projects one committed permission change.
func (broadcaster *Broadcaster) Handle(ctx context.Context, event bus.Event) error {
	payload, ok := event.Payload.(permissionchanged.Payload)
	if !ok {
		return nil
	}
	playerIDs := make([]int64, 0)
	if payload.PlayerID > 0 {
		playerIDs = append(playerIDs, payload.PlayerID)
	}
	if payload.GroupID != nil {
		affected, err := broadcaster.permissions.AffectedPlayerIDs(ctx, *payload.GroupID)
		if err != nil {
			return err
		}
		playerIDs = append(playerIDs, affected...)
	}

	var result error
	seen := make(map[int64]struct{}, len(playerIDs))
	for _, playerID := range playerIDs {
		if _, found := seen[playerID]; found {
			continue
		}
		seen[playerID] = struct{}{}
		connection, found := broadcaster.connection(playerID)
		if !found {
			continue
		}
		packets, err := broadcaster.projector.Packets(ctx, playerID)
		if err != nil {
			result = errors.Join(result, err)
			continue
		}
		for _, packet := range packets {
			result = errors.Join(result, connection.Send(ctx, packet))
		}
	}

	return result
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
