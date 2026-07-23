package games

import (
	"context"
	"time"

	"github.com/niflaot/pixels/internal/realm/room/runtime/broadcast"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	worldfurniture "github.com/niflaot/pixels/internal/realm/room/world/furniture"
	"github.com/niflaot/pixels/internal/realm/room/world/games/freeze"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
	outnumber "github.com/niflaot/pixels/networking/outbound/room/entities/number"
)

// freezeHit stores one successful blast result for projection outside the lock.
type freezeHit struct {
	// playerID identifies the damaged participant.
	playerID int64
	// lives stores the participant's remaining lives.
	lives int
	// defeated reports whether the hit eliminated the participant.
	defeated bool
}

// freezeDrop stores one block transition produced by an explosion.
type freezeDrop struct {
	// item stores the broken block snapshot.
	item worldfurniture.Item
	// power stores the optional reward revealed by the block.
	power freeze.PowerUp
}

// explodeFreeze resolves one server-authored blast.
func (service *Service) explodeFreeze(ctx context.Context, active *roomlive.Room, ball freeze.Throw, now time.Time) error {
	points := freeze.Explosion(ball.Center, ball.Radius, ball.Diagonal, ball.Massive, service.freezePointValidator(active))
	hitPoints := make(map[grid.Point]struct{}, len(points))
	blocks := make([]worldfurniture.Item, 0)
	for _, point := range points {
		hitPoints[point] = struct{}{}
		for _, item := range active.FurnitureAt(point) {
			if item.Definition.InteractionType == "freeze_block" {
				blocks = append(blocks, item)
			}
		}
	}
	hits, drops, matchOver, startedAt := service.resolveFreezeBlast(ctx, active, ball, now, hitPoints, blocks)
	for _, hit := range hits {
		unit, err := active.SetUnitControl(hit.playerID, worldunit.ControlFrozen)
		if err == nil {
			_ = broadcast.RoomUnitStatus(ctx, service.connections, active, unit, 0)
			_ = service.projectFreezeLives(ctx, active, unit.EntityKey, hit.lives)
		}
		if hit.defeated {
			service.teleportFreezeExit(ctx, active, hit.playerID)
		}
	}
	if err := service.animateFreezeTiles(ctx, active, ball, points); err != nil {
		return err
	}
	service.scheduleFreezeDrops(active, ball.Center, drops, startedAt)
	if matchOver {
		return service.finish(ctx, active, startedAt)
	}
	return nil
}

// resolveFreezeBlast mutates participants and reserves newly broken blocks.
func (service *Service) resolveFreezeBlast(ctx context.Context, active *roomlive.Room, ball freeze.Throw, now time.Time, hitPoints map[grid.Point]struct{}, blocks []worldfurniture.Item) ([]freezeHit, []freezeDrop, bool, time.Time) {
	service.mutex.Lock()
	defer service.mutex.Unlock()
	state := service.states[active.ID()]
	if state == nil {
		return nil, nil, false, time.Time{}
	}
	owner := state.freezePlayers[ball.OwnerID]
	hits := make([]freezeHit, 0)
	for _, presence := range active.Presences() {
		player := state.freezePlayers[presence.Occupant.PlayerID]
		if player == nil || player.ID == ball.OwnerID {
			continue
		}
		if _, found := hitPoints[presence.Unit.Position.Point]; !found || !player.Hit(now, service.config.Freeze.FrozenDuration, service.config.Freeze.LooseSnowballs, service.config.Freeze.LooseBoost) {
			continue
		}
		ownerTeam := uint8(0)
		if owner != nil {
			ownerTeam = owner.Team
		}
		service.coordinator.AddScore(ctx, active.ID(), ball.OwnerID, int64(freeze.FreezePoints(ownerTeam, player.Team, service.config.Freeze.PointsFreeze)))
		if ownerTeam != player.Team {
			service.progress(ctx, ball.OwnerID, "game.freeze.player.frozen", 1)
		}
		service.wired.ProjectEffect(active.ID(), player.ID, 12)
		hits = append(hits, freezeHit{playerID: player.ID, lives: player.Lives, defeated: !player.Alive()})
	}
	drops := service.reserveFreezeDrops(state, ball.OwnerID, blocks)
	if owner != nil {
		owner.Snowballs++
		service.coordinator.AddScore(ctx, active.ID(), ball.OwnerID, int64(service.config.Freeze.PointsBlock))
	}
	return hits, drops, freezeMatchOver(state.freezePlayers), state.startedAt
}

// reserveFreezeDrops marks intact blocks broken and selects optional rewards.
func (service *Service) reserveFreezeDrops(state *roomState, ownerID int64, blocks []worldfurniture.Item) []freezeDrop {
	drops := make([]freezeDrop, 0, len(blocks))
	for _, block := range blocks {
		if block.ExtraData != "" && block.ExtraData != "0" {
			continue
		}
		if _, broken := state.freezeDrops[block.ID]; broken {
			continue
		}
		power, found := freeze.Drop(block.ID, ownerID, service.config.Freeze.PowerupChance)
		if !found {
			power = 0
		}
		state.freezeDrops[block.ID] = power
		drops = append(drops, freezeDrop{item: block, power: power})
	}
	return drops
}

// animateFreezeTiles projects the blast wave only on floor tiles.
func (service *Service) animateFreezeTiles(ctx context.Context, active *roomlive.Room, ball freeze.Throw, points []grid.Point) error {
	updates := make([]stateUpdate, 0, len(points))
	for _, point := range points {
		for _, item := range active.FurnitureAt(point) {
			if item.Definition.InteractionType == "freeze_tile" {
				updates = append(updates, stateUpdate{item.ID, freeze.ArmedState(ball.Radius)})
			}
		}
	}
	if len(updates) == 0 {
		return nil
	}
	if err := service.projectStatesBatch(ctx, active, updates); err != nil {
		return err
	}
	for index := range updates {
		item, _ := active.FurnitureItem(updates[index].id)
		updates[index].value = freeze.ResetState(freeze.Distance(ball.Center, item.Point))
	}
	return service.projectStatesBatch(ctx, active, updates)
}

// freezePointValidator returns a surface-bound blast predicate.
func (service *Service) freezePointValidator(active *roomlive.Room) func(grid.Point) bool {
	width, height, tiles := active.SurfaceHeights()
	return func(point grid.Point) bool {
		index := int(point.Y)*int(width) + int(point.X)
		return point.X < width && point.Y < height && index >= 0 && index < len(tiles) && tiles[index].Valid
	}
}

// projectFreezeLives shows the remaining lives above one damaged avatar.
func (service *Service) projectFreezeLives(ctx context.Context, active *roomlive.Room, unitID int64, lives int) error {
	packet, err := outnumber.Encode(unitID, int32(lives))
	if err != nil {
		return err
	}
	return broadcast.RoomPacket(ctx, service.connections, active, packet, 0)
}

// teleportFreezeExit moves a defeated player to the configured exit.
func (service *Service) teleportFreezeExit(ctx context.Context, active *roomlive.Room, playerID int64) {
	exits := active.FurnitureByInteraction("freeze_exit")
	if len(exits) == 0 {
		return
	}
	moved, err := active.TeleportUnit(playerID, exits[0].Point, worldunit.RotationSouth, false, roomlive.TeleportNear)
	if err == nil {
		_ = broadcast.RoomUnitStatuses(ctx, service.connections, active, []roomlive.UnitSnapshot{moved}, 0)
	}
}
