package games

import (
	"context"
	"strconv"
	"time"

	"github.com/niflaot/pixels/internal/realm/room/runtime/broadcast"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/internal/realm/room/world/games/freeze"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	outstate "github.com/niflaot/pixels/networking/outbound/room/furniture/state"
)

// scheduleFreezeDrops reveals broken blocks after the blast reaches them.
func (service *Service) scheduleFreezeDrops(active *roomlive.Room, center grid.Point, drops []freezeDrop, startedAt time.Time) {
	for _, drop := range drops {
		drop := drop
		delay := time.Duration(freeze.Distance(center, drop.item.Point))*100*time.Millisecond + 25*time.Millisecond
		active.Schedule(delay, func(time.Time) {
			if !service.freezeRoundActive(active.ID(), startedAt) {
				return
			}
			_ = service.projectFreezeBlockState(context.Background(), active, drop.item.ID, freeze.DropState(drop.power))
		})
	}
}

// resetFreezeRound restores intact blocks and shows every participant's starting lives.
func (service *Service) resetFreezeRound(ctx context.Context, active *roomlive.Room) error {
	if len(active.FurnitureByInteraction("freeze_tile")) == 0 {
		return nil
	}
	for _, block := range active.FurnitureByInteraction("freeze_block") {
		if err := service.projectFreezeBlockState(ctx, active, block.ID, 0); err != nil {
			return err
		}
	}
	service.mutex.Lock()
	state := service.states[active.ID()]
	if state == nil {
		service.mutex.Unlock()
		return nil
	}
	lives := make(map[int64]int, len(state.freezePlayers))
	for playerID, player := range state.freezePlayers {
		lives[playerID] = player.Lives
	}
	service.mutex.Unlock()
	for playerID, remaining := range lives {
		if unit, found := active.Unit(playerID); found {
			if err := service.projectFreezeLives(ctx, active, unit.EntityKey, remaining); err != nil {
				return err
			}
		}
	}
	return nil
}

// collectFreezePowerUp applies a visible reward only after a participant walks over it.
func (service *Service) collectFreezePowerUp(ctx context.Context, active *roomlive.Room, playerID int64, blockID int64) error {
	snapshot, found := service.wired.Snapshot(active.ID())
	if !found || !snapshot.Running {
		return nil
	}
	service.mutex.Lock()
	state := service.states[active.ID()]
	if state == nil {
		service.mutex.Unlock()
		return nil
	}
	power, dropped := state.freezeDrops[blockID]
	player := state.freezePlayers[playerID]
	if !dropped || power == 0 || player == nil || !player.Alive() {
		service.mutex.Unlock()
		return nil
	}
	delete(state.freezeDrops, blockID)
	player.ApplyPowerUp(power, service.config.Freeze.MaxSnowballs, service.config.Freeze.MaxLives, time.Now(), service.config.Freeze.ProtectionDuration, service.config.Freeze.ProtectionStack)
	team, lives, startedAt := player.Team, player.Lives, state.startedAt
	service.mutex.Unlock()
	service.coordinator.AddScore(ctx, active.ID(), playerID, int64(service.config.Freeze.PointsEffect))
	service.progress(ctx, playerID, "game.freeze.powerup", 1)
	if power == freeze.Shield {
		service.wired.ProjectEffect(active.ID(), playerID, 48+int32(team))
	}
	if power == freeze.LifeUp {
		if unit, found := active.Unit(playerID); found {
			_ = service.projectFreezeLives(ctx, active, unit.EntityKey, lives)
		}
	}
	if err := service.projectFreezeBlockState(ctx, active, blockID, freeze.CollectedState(power)); err != nil {
		return err
	}
	active.Schedule(750*time.Millisecond, func(time.Time) {
		if service.freezeRoundActive(active.ID(), startedAt) {
			_ = service.projectFreezeBlockState(context.Background(), active, blockID, freeze.DropState(0))
		}
	})
	return nil
}

// freezeRoundActive reports whether scheduled work still belongs to the running round.
func (service *Service) freezeRoundActive(roomID int64, startedAt time.Time) bool {
	snapshot, found := service.wired.Snapshot(roomID)
	if !found || !snapshot.Running {
		return false
	}
	service.mutex.Lock()
	state := service.states[roomID]
	active := state != nil && state.startedAt.Equal(startedAt)
	service.mutex.Unlock()
	return active
}

// projectFreezeBlockState updates one block's rendering and collision atomically.
func (service *Service) projectFreezeBlockState(ctx context.Context, active *roomlive.Room, itemID int64, value int) error {
	if _, err := active.UpdateFurnitureState(itemID, strconv.Itoa(value), true); err != nil {
		return err
	}
	packet, err := outstate.Encode(itemID, value)
	if err != nil {
		return err
	}
	return broadcast.RoomPacket(ctx, service.connections, active, packet, 0)
}
