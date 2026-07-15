package core

import (
	"context"

	botrecord "github.com/niflaot/pixels/internal/realm/bot/record"
	"github.com/niflaot/pixels/internal/realm/room/runtime/broadcast"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	roomprojection "github.com/niflaot/pixels/internal/realm/room/runtime/projection"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	outskilllist "github.com/niflaot/pixels/networking/outbound/bot/skilllist"
	outadd "github.com/niflaot/pixels/networking/outbound/inventory/bots/add"
	outlist "github.com/niflaot/pixels/networking/outbound/inventory/bots/list"
	outremoveinventory "github.com/niflaot/pixels/networking/outbound/inventory/bots/remove"
	outdance "github.com/niflaot/pixels/networking/outbound/room/entities/dance"
	outeffect "github.com/niflaot/pixels/networking/outbound/room/entities/effect"
	outremoved "github.com/niflaot/pixels/networking/outbound/room/entities/removed"
	outstatus "github.com/niflaot/pixels/networking/outbound/room/entities/status"
	outunits "github.com/niflaot/pixels/networking/outbound/room/entities/units"
)

// botSkills returns the Nitro commands supported by one behavior.
func botSkills(bot botrecord.Bot) []uint16 {
	skills := []uint16{0, 1, 2, 3, 4, 5, 9}
	if bot.BehaviorType == botrecord.BehaviorBartender {
		skills = append(skills, 6)
	}
	return skills
}

// inventoryBot maps durable data into Nitro inventory data.
func inventoryBot(bot botrecord.Bot) outlist.Bot {
	gender := bot.Gender
	if gender != "" {
		gender = string(gender[0] | 0x20)
	}
	return outlist.Bot{ID: bot.ID, Name: bot.Name, Motto: bot.Motto, Gender: gender, Figure: bot.Figure}
}

// InventoryRecord maps durable data into Nitro inventory data.
func InventoryRecord(bot botrecord.Bot) outlist.Bot { return inventoryBot(bot) }

// SkillRecords returns the current Nitro skill menu for one bot.
func SkillRecords(bot botrecord.Bot) []outskilllist.Skill {
	ids := botSkills(bot)
	result := make([]outskilllist.Skill, len(ids))
	for index, id := range ids {
		result[index] = outskilllist.Skill{ID: int32(id)}
	}
	return result
}

// roomBot maps one active bot into Nitro's rentable-bot UNIT variant.
func roomBot(bot botrecord.Bot, unit roomlive.UnitSnapshot) outunits.Unit {
	return outunits.Unit{Type: outunits.RentableBotType, UserID: -bot.ID, Name: bot.Name, Motto: bot.Motto, Figure: bot.Figure, RoomIndex: unit.UnitID, X: int32(unit.Position.Point.X), Y: int32(unit.Position.Point.Y), Z: unit.Position.Z.String(), Direction: int32(unit.BodyRotation), Gender: bot.Gender, OwnerID: bot.OwnerPlayerID, OwnerName: bot.OwnerName, Skills: botSkills(bot)}
}

// ProjectSpawn broadcasts one bot spawn and persistent visual state.
func (service *Service) ProjectSpawn(ctx context.Context, active *roomlive.Room, bot botrecord.Bot) {
	unit, found := active.Unit(EntityKey(bot.ID))
	if !found {
		return
	}
	packet, err := outunits.Encode([]outunits.Unit{roomBot(bot, unit)})
	if err == nil {
		_ = broadcast.RoomPacket(ctx, service.connections, active, packet, 0)
	}
	service.projectStatus(ctx, active, bot, unit, nil)
}

// projectSpawnConnection sends one bot snapshot to a late room entrant.
func (service *Service) projectSpawnConnection(ctx context.Context, connection netconn.Connection, active *roomlive.Room, bot botrecord.Bot) {
	unit, found := active.Unit(EntityKey(bot.ID))
	if !found || connection == nil {
		return
	}
	packets := make([]codec.Packet, 0, 4)
	if packet, err := outunits.Encode([]outunits.Unit{roomBot(bot, unit)}); err == nil {
		packets = append(packets, packet)
	}
	if statuses := roomprojection.Statuses(active, EntityKey(bot.ID)); len(statuses) > 0 {
		if packet, err := outstatus.Encode(statuses); err == nil {
			packets = append(packets, packet)
		}
	}
	packets = append(packets, persistentPackets(bot, unit)...)
	for _, packet := range packets {
		_ = connection.Send(ctx, packet)
	}
}

// projectStatus broadcasts current bot status and optional movement-independent visuals.
func (service *Service) projectStatus(ctx context.Context, active *roomlive.Room, bot botrecord.Bot, unit roomlive.UnitSnapshot, statusPacket *codec.Packet) {
	if statusPacket != nil {
		_ = broadcast.RoomPacket(ctx, service.connections, active, *statusPacket, 0)
	} else if statuses := roomprojection.Statuses(active, EntityKey(bot.ID)); len(statuses) > 0 {
		if packet, err := outstatus.Encode(statuses); err == nil {
			_ = broadcast.RoomPacket(ctx, service.connections, active, packet, 0)
		}
	}
	for _, packet := range persistentPackets(bot, unit) {
		_ = broadcast.RoomPacket(ctx, service.connections, active, packet, 0)
	}
}

// persistentPackets encodes selected dance and effect state.
func persistentPackets(bot botrecord.Bot, unit roomlive.UnitSnapshot) []codec.Packet {
	packets := make([]codec.Packet, 0, 2)
	if bot.DanceType > 0 {
		if packet, err := outdance.Encode(unit.UnitID, int32(bot.DanceType)); err == nil {
			packets = append(packets, packet)
		}
	}
	if bot.EffectID != nil && *bot.EffectID > 0 {
		if packet, err := outeffect.Encode(unit.UnitID, *bot.EffectID, 0); err == nil {
			packets = append(packets, packet)
		}
	}
	return packets
}

// ProjectRemove broadcasts one room bot removal.
func (service *Service) ProjectRemove(ctx context.Context, active *roomlive.Room, unitID int64) {
	packet, err := outremoved.Encode(unitID)
	if err == nil {
		_ = broadcast.RoomPacket(ctx, service.connections, active, packet, 0)
	}
}

// SendInventoryRemove notifies one online owner that a placed bot left inventory.
func (service *Service) SendInventoryRemove(ctx context.Context, playerID int64, botID int64) {
	packet, err := outremoveinventory.Encode(botID)
	if err == nil {
		service.sendPlayer(ctx, playerID, packet)
	}
}

// SendInventoryAdd notifies one online owner that a bot entered inventory.
func (service *Service) SendInventoryAdd(ctx context.Context, playerID int64, bot botrecord.Bot, open bool) {
	packet, err := outadd.Encode(inventoryBot(bot), open)
	if err == nil {
		service.sendPlayer(ctx, playerID, packet)
	}
}

// sendPlayer performs one best-effort live player delivery.
func (service *Service) sendPlayer(ctx context.Context, playerID int64, packet codec.Packet) {
	connection, found := service.playerConnection(playerID)
	if found {
		_ = connection.Send(ctx, packet)
	}
}

// playerConnection resolves one live player's active connection.
func (service *Service) playerConnection(playerID int64) (netconn.Connection, bool) {
	player, found := service.players.Find(playerID)
	if !found {
		return nil, false
	}
	peer := player.Peer()
	return service.connections.Get(peer.ConnectionKind(), peer.ConnectionID())
}
