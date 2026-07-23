package talent

import (
	"context"

	permissionbroadcast "github.com/niflaot/pixels/internal/permission/broadcast"
	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	progressionrecord "github.com/niflaot/pixels/internal/realm/progression/record"
	netconn "github.com/niflaot/pixels/networking/connection"
	talentdata "github.com/niflaot/pixels/networking/outbound/progression/talent/data"
	outlevelup "github.com/niflaot/pixels/networking/outbound/progression/talent/levelup"
)

// LiveProjector publishes paid talent levels and refreshed client perks.
type LiveProjector struct {
	// players resolves online players.
	players *playerlive.Registry
	// connections sends client packets.
	connections *netconn.Registry
	// permissions builds current client perk snapshots.
	permissions *permissionbroadcast.Projector
	// furniture resolves product names and sprites.
	furniture *furnitureservice.Service
}

// NewLiveProjector creates one talent projection service.
func NewLiveProjector(players *playerlive.Registry, connections *netconn.Registry, permissions *permissionbroadcast.Projector, furniture *furnitureservice.Service) *LiveProjector {
	return &LiveProjector{players: players, connections: connections, permissions: permissions, furniture: furniture}
}

// LevelUp publishes one newly paid talent level.
func (projector *LiveProjector) LevelUp(ctx context.Context, playerID int64, level progressionrecord.TalentLevel) {
	connection, found := projector.connection(playerID)
	if !found {
		return
	}
	perks := make([]int32, 0, len(level.RewardPerks))
	for index := range level.RewardPerks {
		perks = append(perks, int32(index+1))
	}
	products := projector.products(ctx, level.RewardItems)
	if len(perks) > 0 {
		if packet, err := outlevelup.Encode(level.Track, level.Level, perks, nil); err == nil {
			_ = connection.Send(ctx, packet)
		}
	}
	if len(products) > 0 {
		if packet, err := outlevelup.Encode(level.Track, level.Level, nil, products); err == nil {
			_ = connection.Send(ctx, packet)
		}
	}
	if len(perks) == 0 && len(products) == 0 {
		if packet, err := outlevelup.Encode(level.Track, level.Level, nil, nil); err == nil {
			_ = connection.Send(ctx, packet)
		}
	}
	if projector.permissions != nil {
		packets, err := projector.permissions.Packets(ctx, playerID)
		if err == nil {
			for _, packet := range packets {
				_ = connection.Send(ctx, packet)
			}
		}
	}
}

// products resolves configured furniture rewards into client product metadata.
func (projector *LiveProjector) products(ctx context.Context, ids []int64) []talentdata.Product {
	values := make([]talentdata.Product, 0, len(ids))
	if projector.furniture == nil {
		return values
	}
	for _, id := range ids {
		definition, found, err := projector.furniture.FindDefinitionByID(ctx, id)
		if err == nil && found {
			values = append(values, talentdata.Product{Name: definition.Name, Value: int32(definition.SpriteID)})
		}
	}
	return values
}

// connection resolves one online player's active connection.
func (projector *LiveProjector) connection(playerID int64) (netconn.Connection, bool) {
	if projector == nil || projector.players == nil || projector.connections == nil {
		return nil, false
	}
	player, found := projector.players.Find(playerID)
	if !found {
		return nil, false
	}
	peer := player.Peer()
	return projector.connections.Get(peer.ConnectionKind(), peer.ConnectionID())
}
