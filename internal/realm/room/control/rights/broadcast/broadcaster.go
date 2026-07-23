// Package broadcast projects committed room rights into active rooms and clients.
package broadcast

import (
	"context"

	playerservice "github.com/niflaot/pixels/internal/realm/player/service"
	"github.com/niflaot/pixels/internal/realm/room/runtime/broadcast"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/internal/realm/room/runtime/projection"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	outstatus "github.com/niflaot/pixels/networking/outbound/room/entities/status"
	outadded "github.com/niflaot/pixels/networking/outbound/room/rights/added"
	outlevel "github.com/niflaot/pixels/networking/outbound/room/rights/level"
	outremoved "github.com/niflaot/pixels/networking/outbound/room/rights/removed"
)

// Broadcaster projects room rights mutations.
type Broadcaster struct {
	// runtime stores active rooms.
	runtime *roomlive.Registry
	// connections stores active connections.
	connections *netconn.Registry
	// players resolves durable usernames.
	players playerservice.Finder
}

// New creates a room rights broadcaster.
func New(runtime *roomlive.Registry, connections *netconn.Registry, players playerservice.Finder) *Broadcaster {
	return &Broadcaster{runtime: runtime, connections: connections, players: players}
}

// Granted projects one committed rights grant.
func (broadcaster *Broadcaster) Granted(ctx context.Context, roomID int64, playerID int64) error {
	active, found := broadcaster.runtime.Find(roomID)
	if !found {
		return nil
	}
	active.GrantRights(playerID)
	record, found, err := broadcaster.players.FindByID(ctx, playerID)
	if err != nil || !found {
		return err
	}
	packet, err := outadded.Encode(int32(roomID), int32(playerID), record.Player.Username)
	if err != nil {
		return err
	}
	if err := broadcast.RoomPacket(ctx, broadcaster.connections, active, packet, 0); err != nil {
		return err
	}
	active.SetUnitStatus(playerID, worldunit.StatusFlatControl, "1")
	if err := broadcaster.broadcastStatus(ctx, active, playerID); err != nil {
		return err
	}

	return broadcaster.sendLevel(ctx, active, playerID, outlevel.Rights)
}

// Revoked projects one committed rights revocation.
func (broadcaster *Broadcaster) Revoked(ctx context.Context, roomID int64, playerID int64) error {
	active, found := broadcaster.runtime.Find(roomID)
	if !found {
		return nil
	}
	active.RevokeRights(playerID)
	packet, err := outremoved.Encode(int32(roomID), int32(playerID))
	if err != nil {
		return err
	}
	if err := broadcast.RoomPacket(ctx, broadcaster.connections, active, packet, 0); err != nil {
		return err
	}
	active.SetUnitStatus(playerID, worldunit.StatusFlatControl, "0")
	if err := broadcaster.broadcastStatus(ctx, active, playerID); err != nil {
		return err
	}

	return broadcaster.sendLevel(ctx, active, playerID, outlevel.None)
}

// broadcastStatus synchronizes one target's room controller marker with every occupant.
func (broadcaster *Broadcaster) broadcastStatus(ctx context.Context, active *roomlive.Room, playerID int64) error {
	records := projection.Statuses(active, playerID)
	if len(records) == 0 {
		return nil
	}
	packet, err := outstatus.Encode(records)
	if err != nil {
		return err
	}

	return broadcast.RoomPacket(ctx, broadcaster.connections, active, packet, 0)
}

// sendLevel sends room control level to one active occupant.
func (broadcaster *Broadcaster) sendLevel(ctx context.Context, active *roomlive.Room, playerID int64, value int32) error {
	packet, err := outlevel.Encode(value)
	if err != nil {
		return err
	}

	return sendToOccupant(ctx, broadcaster.connections, active, playerID, packet)
}

// sendToOccupant sends one packet to a matching active occupant.
func sendToOccupant(ctx context.Context, connections *netconn.Registry, active *roomlive.Room, playerID int64, packet codec.Packet) error {
	if connections == nil {
		return nil
	}
	for _, occupant := range active.Occupants() {
		if occupant.PlayerID != playerID {
			continue
		}
		connection, found := connections.Get(occupant.ConnectionKind, occupant.ConnectionID)
		if !found {
			return nil
		}

		return connection.Send(ctx, packet)
	}

	return nil
}
