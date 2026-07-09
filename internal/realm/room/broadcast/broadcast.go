// Package broadcast sends room runtime packets to active connections.
package broadcast

import (
	"context"

	"github.com/niflaot/pixels/internal/realm/room/live"
	"github.com/niflaot/pixels/internal/realm/room/projection"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	outremoved "github.com/niflaot/pixels/networking/outbound/room/entities/removed"
	outstatus "github.com/niflaot/pixels/networking/outbound/room/entities/status"
	outunits "github.com/niflaot/pixels/networking/outbound/room/entities/units"
	outheightmapupdate "github.com/niflaot/pixels/networking/outbound/room/heightmapupdate"
)

// NewMovementPublisher creates a movement broadcaster.
func NewMovementPublisher(connections *netconn.Registry) live.MovementPublisher {
	return func(ctx context.Context, active *live.Room, movements []live.Movement) error {
		if connections == nil || active == nil || len(movements) == 0 {
			return nil
		}

		packet, err := outstatus.Encode(projection.MovementStatuses(movements))
		if err != nil {
			return err
		}

		return RoomPacket(ctx, connections, active, packet, 0)
	}
}

// RoomPacket sends a packet to active room occupants. Delivery is best-effort: a failed send to
// one occupant (typically a connection mid-disconnect) never fails the caller, because the failing
// connection's own lifecycle handles its cleanup and a command must not disconnect the acting
// player just because a bystander's socket died.
func RoomPacket(ctx context.Context, connections *netconn.Registry, active *live.Room, packet codec.Packet, excludedPlayerID int64) error {
	if connections == nil || active == nil {
		return nil
	}

	for _, occupant := range active.Occupants() {
		if occupant.PlayerID == excludedPlayerID {
			continue
		}
		connection, found := connections.Get(occupant.ConnectionKind, occupant.ConnectionID)
		if !found {
			continue
		}
		_ = connection.Send(ctx, packet)
	}

	return nil
}

// RoomSpawn sends a unit spawn and initial status to active room occupants.
func RoomSpawn(ctx context.Context, connections *netconn.Registry, active *live.Room, playerID int64, excludedPlayerID int64) error {
	unitRecords := projection.Units(active, playerID)
	if len(unitRecords) > 0 {
		packet, err := outunits.Encode(unitRecords)
		if err != nil {
			return err
		}
		if err := RoomPacket(ctx, connections, active, packet, excludedPlayerID); err != nil {
			return err
		}
	}

	statusRecords := projection.Statuses(active, playerID)
	if len(statusRecords) == 0 {
		return nil
	}
	packet, err := outstatus.Encode(statusRecords)
	if err != nil {
		return err
	}

	return RoomPacket(ctx, connections, active, packet, excludedPlayerID)
}

// RoomUnitStatus sends one unit status snapshot to active room occupants.
func RoomUnitStatus(ctx context.Context, connections *netconn.Registry, active *live.Room, unit live.UnitSnapshot, excludedPlayerID int64) error {
	return RoomUnitStatuses(ctx, connections, active, []live.UnitSnapshot{unit}, excludedPlayerID)
}

// RoomUnitStatuses sends several unit status snapshots to active room occupants in one packet, doing
// nothing when there is nothing to report (e.g. a furniture change with no reoriented occupants).
func RoomUnitStatuses(ctx context.Context, connections *netconn.Registry, active *live.Room, units []live.UnitSnapshot, excludedPlayerID int64) error {
	if len(units) == 0 {
		return nil
	}

	movements := make([]live.Movement, 0, len(units))
	for _, unit := range units {
		movements = append(movements, live.Movement{Unit: unit, Settled: true})
	}

	packet, err := outstatus.Encode(projection.MovementStatuses(movements))
	if err != nil {
		return err
	}

	return RoomPacket(ctx, connections, active, packet, excludedPlayerID)
}

// RoomRemove sends a room unit remove packet to room occupants.
func RoomRemove(ctx context.Context, connections *netconn.Registry, active *live.Room, unitID int64, excludedPlayerID int64) error {
	packet, err := outremoved.Encode(unitID)
	if err != nil {
		return err
	}

	return RoomPacket(ctx, connections, active, packet, excludedPlayerID)
}

// RoomHeightMapUpdate sends the current tile heights at specific points to active room occupants,
// keeping each client's cached local height map (used for placement and movement prediction) in
// sync after furniture is placed, moved, rotated, or picked up. Does nothing when active has no
// loaded world or none of the requested points resolve to a room tile.
func RoomHeightMapUpdate(ctx context.Context, connections *netconn.Registry, active *live.Room, points []grid.Point, excludedPlayerID int64) error {
	if active == nil || len(points) == 0 {
		return nil
	}

	width, _, tiles := active.SurfaceHeights()
	records := projection.HeightMapUpdateTiles(width, tiles, points)
	if len(records) == 0 {
		return nil
	}

	packet, err := outheightmapupdate.Encode(records)
	if err != nil {
		return err
	}

	return RoomPacket(ctx, connections, active, packet, excludedPlayerID)
}
