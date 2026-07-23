package games

import (
	"context"

	"github.com/niflaot/pixels/internal/realm/room/runtime/broadcast"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
)

// cleanupFreezeMatch removes every temporary Freeze projection and pending action.
func (service *Service) cleanupFreezeMatch(ctx context.Context, active *roomlive.Room) {
	service.mutex.Lock()
	state := service.states[active.ID()]
	if state == nil {
		service.mutex.Unlock()
		return
	}
	playerIDs := make([]int64, 0, len(state.freezePlayers))
	for playerID := range state.freezePlayers {
		playerIDs = append(playerIDs, playerID)
	}
	state.freezeBalls = state.freezeBalls[:0]
	clear(state.freezePlayers)
	clear(state.freezeDrops)
	service.mutex.Unlock()
	for _, playerID := range playerIDs {
		service.wired.ProjectEffect(active.ID(), playerID, 0)
		if unit, err := active.ReleaseUnitControl(playerID); err == nil {
			_ = broadcast.RoomUnitStatus(ctx, service.connections, active, unit, 0)
		}
		_ = service.sendPlaying(ctx, active, playerID, false)
	}
	for _, block := range active.FurnitureByInteraction("freeze_block") {
		_ = service.projectFreezeBlockState(ctx, active, block.ID, 0)
	}
}
