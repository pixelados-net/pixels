package runtime

import (
	"context"
	"strconv"
	"time"

	groupobservability "github.com/niflaot/pixels/internal/realm/group/observability"
	grouprecord "github.com/niflaot/pixels/internal/realm/group/record"
	roombroadcast "github.com/niflaot/pixels/internal/realm/room/runtime/broadcast"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	netconn "github.com/niflaot/pixels/networking/connection"
	"github.com/niflaot/pixels/networking/outbound/furniture/stuffdata"
	outbadges "github.com/niflaot/pixels/networking/outbound/group/badge/list"
	outfavorite "github.com/niflaot/pixels/networking/outbound/group/membership/favorite/update"
	outobjectbatch "github.com/niflaot/pixels/networking/outbound/room/furniture/objectdata/batch"
)

// Projector updates current-room social-group unit state without respawning units.
type Projector struct {
	// cache resolves immutable player generations.
	cache *Cache
	// rooms resolves the player's active room.
	rooms *roomlive.Registry
	// connections sends room projections.
	connections *netconn.Registry
	// delivery sends inventory projections to affected online players.
	delivery *Delivery
	// metrics stores bounded process-wide group telemetry.
	metrics *groupobservability.Metrics
}

// GroupFurnitureColors projects current colors to linked items in active rooms.
func (projector *Projector) GroupFurnitureColors(ctx context.Context, groupID int64) error {
	if projector == nil || projector.rooms == nil || projector.cache == nil || projector.connections == nil {
		return nil
	}
	started := time.Now()
	defer func() { projector.metrics.Observe(groupobservability.ProjectionFanout, time.Since(started)) }()
	group, loaded := projector.cache.Group(groupID)
	if !loaded {
		return nil
	}
	groupIDValue := strconv.FormatInt(groupID, 10)
	for _, room := range projector.rooms.Snapshot() {
		itemIDs := make([]int64, 0)
		data := make([]*stuffdata.Data, 0)
		for _, item := range room.FurnitureItems() {
			linkedGroupID, found := projector.cache.FurnitureGroup(item.ID)
			if !found || linkedGroupID != groupID {
				continue
			}
			values := []string{item.ExtraData, groupIDValue, group.Group.BadgeCode, group.Group.ColorAHex, group.Group.ColorBHex}
			itemIDs = append(itemIDs, item.ID)
			data = append(data, stuffdata.StringArray(values))
		}
		if len(itemIDs) == 0 {
			continue
		}
		packet, err := outobjectbatch.Encode(itemIDs, data)
		if err != nil {
			return err
		}
		if err = roombroadcast.RoomPacket(ctx, projector.connections, room, packet, 0); err != nil {
			return err
		}
	}
	return nil
}

// SetMetrics attaches process-wide telemetry before serving requests.
func (projector *Projector) SetMetrics(metrics *groupobservability.Metrics) {
	projector.metrics = metrics
}

// NewProjector creates live social-group projection behavior.
func NewProjector(cache *Cache, rooms *roomlive.Registry, connections *netconn.Registry, delivery *Delivery) *Projector {
	return &Projector{cache: cache, rooms: rooms, connections: connections, delivery: delivery}
}

// Favorite projects one player's current favorite to every current-room occupant.
func (projector *Projector) Favorite(ctx context.Context, playerID int64) error {
	if projector == nil || projector.rooms == nil || projector.cache == nil {
		return nil
	}
	room, found := projector.rooms.FindByPlayer(playerID)
	if !found {
		return nil
	}
	groupID, status, name := int64(-1), int32(-1), ""
	var badges []grouprecord.PlayerGroup
	if player, loaded := projector.cache.Player(playerID); loaded {
		if group, role, favorite := player.Favorite(); favorite {
			groupID, status, name = group.ID, int32(role), group.Name
			badges = []grouprecord.PlayerGroup{{Group: group, Role: role, Favorite: true}}
		}
	}
	if !room.UpdateOccupantGroup(playerID, groupID, status, name) {
		return nil
	}
	unit, found := room.Unit(playerID)
	if !found {
		return nil
	}
	packet, err := outfavorite.Encode(int32(unit.UnitID), groupID, status, name)
	if err != nil {
		return err
	}
	if err = roombroadcast.RoomPacket(ctx, projector.connections, room, packet, 0); err != nil {
		return err
	}
	if len(badges) == 0 {
		return nil
	}
	badgePacket, err := outbadges.Encode(badges)
	if err != nil {
		return err
	}
	return roombroadcast.RoomPacket(ctx, projector.connections, room, badgePacket, 0)
}
