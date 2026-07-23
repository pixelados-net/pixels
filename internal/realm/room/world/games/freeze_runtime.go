package games

import (
	"context"
	"time"

	"github.com/niflaot/pixels/internal/realm/room/runtime/broadcast"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/internal/realm/room/world/games/freeze"
)

// initializeMatchLocked initializes game-specific state under the service lock.
func (service *Service) initializeMatchLocked(active *roomlive.Room, state *roomState, now time.Time) {
	clear(state.freezePlayers)
	clear(state.freezeDrops)
	state.freezeBalls = state.freezeBalls[:0]
	state.nextTagCredit = now.Add(time.Minute)
	if len(active.FurnitureByInteraction("freeze_tile")) == 0 {
		return
	}
	snapshot, _ := service.wired.Snapshot(active.ID())
	for playerID, team := range snapshot.Teams {
		state.freezePlayers[playerID] = &freeze.Player{ID: playerID, Team: uint8(team), Lives: service.config.Freeze.MaxLives, Snowballs: 1, Radius: 1}
		service.wired.ProjectEffect(active.ID(), playerID, 39+team)
	}
}

// throwFreeze schedules one explosion on the existing room tick.
func (service *Service) throwFreeze(ctx context.Context, request UseRequest) error {
	snapshot, found := service.wired.Snapshot(request.Room.ID())
	if !found || !snapshot.Running {
		return nil
	}
	blockID := request.blockID
	if request.Item.Definition.InteractionType == "freeze_block" {
		blockID = request.Item.ID
		request.blockID = blockID
		for _, item := range request.Room.FurnitureAt(request.Item.Point) {
			if item.Definition.InteractionType == "freeze_tile" {
				request.Item = item
				break
			}
		}
		if request.Item.ID == blockID {
			return nil
		}
	}
	unit, present := request.Room.Unit(request.PlayerID)
	if !present || !pointsAdjacent(unit.Position.Point, request.Item.Point) {
		if present {
			for _, point := range freeze.ApproachPoints(request.Item.Point, unit.Position.Point) {
				roomPath, moveErr := request.Room.MoveTo(request.PlayerID, point)
				if moveErr != nil || roomPath.Len() == 0 {
					continue
				}
				request.Room.Schedule(time.Duration(roomPath.Len())*roomlive.DefaultTickInterval, func(time.Time) {
					_ = service.throwFreeze(context.Background(), request)
				})
				break
			}
		}
		return nil
	}
	service.mutex.Lock()
	state := service.stateLocked(request.Room)
	player := state.freezePlayers[request.PlayerID]
	_, blockBroken := state.freezeDrops[blockID]
	if player == nil || blockBroken || !player.Alive() || player.Snowballs < 1 || player.FrozenUntil.After(time.Now()) {
		service.mutex.Unlock()
		return nil
	}
	player.Snowballs--
	state.freezeBalls = append(state.freezeBalls, freeze.Throw{OwnerID: request.PlayerID, ItemID: request.Item.ID, BlockID: blockID, Center: request.Item.Point, Deadline: time.Now().Add(2 * time.Second), Radius: player.Radius, Diagonal: player.Diagonal, Massive: player.Massive})
	service.metrics.freezeBalls.Add(1)
	player.Diagonal, player.Massive = false, false
	radius := player.Radius
	service.mutex.Unlock()
	return service.projectState(ctx, request.Room, request.Item.ID, freeze.ArmedState(radius))
}

// cycleFreeze executes due explosions and restores expired effects.
func (service *Service) cycleFreeze(ctx context.Context, active *roomlive.Room, now time.Time) error {
	service.mutex.Lock()
	state := service.states[active.ID()]
	if state == nil || len(state.freezePlayers) == 0 {
		service.mutex.Unlock()
		return nil
	}
	due := make([]freeze.Throw, 0, len(state.freezeBalls))
	released := make([]int64, 0)
	remaining := state.freezeBalls[:0]
	for _, ball := range state.freezeBalls {
		if !ball.Deadline.After(now) {
			due = append(due, ball)
		} else {
			remaining = append(remaining, ball)
		}
	}
	state.freezeBalls = remaining
	for playerID, player := range state.freezePlayers {
		if !player.FrozenUntil.IsZero() && !player.FrozenUntil.After(now) {
			player.FrozenUntil = time.Time{}
			released = append(released, playerID)
			service.wired.ProjectEffect(active.ID(), playerID, 39+int32(player.Team))
		}
	}
	service.mutex.Unlock()
	for _, playerID := range released {
		unit, err := active.ReleaseUnitControl(playerID)
		if err == nil {
			_ = broadcast.RoomUnitStatus(ctx, service.connections, active, unit, 0)
		}
	}
	for _, ball := range due {
		if err := service.explodeFreeze(ctx, active, ball, now); err != nil {
			return err
		}
	}
	return nil
}
