package game

import (
	"context"

	"github.com/niflaot/pixels/internal/realm/room/runtime/broadcast"
	roomprojection "github.com/niflaot/pixels/internal/realm/room/runtime/projection"
	outeffect "github.com/niflaot/pixels/networking/outbound/room/entities/effect"
)

const teamEffectBase int32 = 32

// projectTeam stores and broadcasts Battle Banzai-compatible team colors.
func (service *Service) projectTeam(roomID int64, playerID int64, team int32) {
	if service.roomWorlds == nil {
		return
	}
	active, found := service.roomWorlds.Find(roomID)
	if !found {
		return
	}
	effectID := int32(0)
	if team >= 1 && team <= 4 {
		effectID = teamEffectBase + team
	}
	unit, found := active.SetUnitEffect(playerID, effectID)
	if !found {
		return
	}
	packet, err := outeffect.Encode(unit.UnitID, roomprojection.EffectID(unit), 0)
	if err == nil {
		_ = broadcast.RoomPacket(context.Background(), service.connections, active, packet, 0)
	}
}

// ProjectEffect stores and broadcasts one room-scoped game effect.
func (service *Service) ProjectEffect(roomID int64, playerID int64, effectID int32) {
	if service.roomWorlds == nil {
		return
	}
	active, found := service.roomWorlds.Find(roomID)
	if !found {
		return
	}
	unit, found := active.SetUnitEffect(playerID, effectID)
	if !found {
		return
	}
	packet, err := outeffect.Encode(unit.UnitID, roomprojection.EffectID(unit), 0)
	if err == nil {
		_ = broadcast.RoomPacket(context.Background(), service.connections, active, packet, 0)
	}
}
