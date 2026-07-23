package games

import (
	"context"
	"strconv"
	"strings"

	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
	"github.com/niflaot/pixels/internal/realm/room/runtime/broadcast"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	worldfurniture "github.com/niflaot/pixels/internal/realm/room/world/furniture"
	"github.com/niflaot/pixels/internal/realm/room/world/games/football"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	rolling "github.com/niflaot/pixels/networking/outbound/room/furniture/rolling"
)

// blockedFootball reports an occupied, non-stackable, or closed goal tile.
func (service *Service) blockedFootball(active *roomlive.Room, point grid.Point, sourceZ float64, direction uint8) bool {
	width, height, tiles := active.SurfaceHeights()
	index := int(point.Y)*int(width) + int(point.X)
	if point.X >= width || point.Y >= height || index < 0 || index >= len(tiles) || !tiles[index].Valid || tiles[index].Height.Units()-sourceZ > 1.65 {
		return true
	}
	if active.HasUnitAt(point) {
		return true
	}
	for _, item := range active.FurnitureAt(point) {
		kind := item.Definition.InteractionType
		if kind == "football" || kind == "battlebanzai_puck" {
			continue
		}
		if strings.HasPrefix(kind, "football_goal_") {
			if !football.GoalScores(direction, uint8(item.Rotation)) {
				return true
			}
			continue
		}
		if !item.Definition.AllowStack {
			return true
		}
	}
	return false
}

// resolveFootballMove chooses the first valid forward or rebound tile.
func (service *Service) resolveFootballMove(active *roomlive.Room, item worldfurniture.Item, direction uint8) (grid.Point, uint8, bool) {
	target, valid := footballPoint(item.Point, direction)
	if valid && !service.blockedFootball(active, target, item.Z.Units(), direction) {
		return target, direction, true
	}
	previous := uint8(255)
	for _, rebound := range football.Rebounds(direction) {
		if rebound == previous {
			continue
		}
		previous = rebound
		target, valid = footballPoint(item.Point, rebound)
		if valid && !service.blockedFootball(active, target, item.Z.Units(), rebound) {
			return target, rebound, true
		}
	}
	return grid.Point{}, direction, false
}

// footballPoint resolves one adjacent target without coordinate wrapping.
func footballPoint(origin grid.Point, direction uint8) (grid.Point, bool) {
	vector := football.Direction(direction)
	return grid.NewPoint(int(origin.X)+vector.X, int(origin.Y)+vector.Y)
}

// moveFootball persists and projects one authoritative ball movement.
func (service *Service) moveFootball(ctx context.Context, active *roomlive.Room, item worldfurniture.Item, target grid.Point) error {
	_, err := service.furniture.Move(ctx, furnitureservice.MoveParams{ItemID: item.ID, ActorPlayerID: item.OwnerPlayerID, RoomID: active.ID(), Placement: furnituremodel.Placement{X: int(target.X), Y: int(target.Y), Z: item.Z.Units(), Rotation: furnituremodel.Rotation(item.Rotation)}})
	if err != nil {
		return err
	}
	previous := item.Point
	item.Point = target
	if _, err = active.ReloadFurniture(item.ID, &item); err != nil {
		return err
	}
	packet, err := rolling.Encode(int(previous.X), int(previous.Y), int(target.X), int(target.Y), []rolling.Item{{ID: item.ID, FromZ: item.Z.String(), ToZ: item.Z.String()}}, item.ID)
	if err != nil {
		return err
	}
	return broadcast.RoomPacket(ctx, service.connections, active, packet, 0)
}

// resetFootball returns a scored ball to its room-load kickoff tile.
func (service *Service) resetFootball(ctx context.Context, active *roomlive.Room, itemID int64) error {
	service.mutex.Lock()
	state := service.states[active.ID()]
	if state == nil {
		service.mutex.Unlock()
		return nil
	}
	origin, found := state.footballOrigins[itemID]
	service.mutex.Unlock()
	if !found {
		return nil
	}
	item, present := active.FurnitureItem(itemID)
	var err error
	if present && item.Point != origin && service.furniture != nil {
		err = service.moveFootball(ctx, active, item, origin)
	}
	service.completeFootballReset(active, itemID)
	return err
}

// completeFootballReset releases one scored ball for its next kick.
func (service *Service) completeFootballReset(active *roomlive.Room, itemID int64) {
	service.mutex.Lock()
	if state := service.states[active.ID()]; state != nil {
		if ball := state.footballs[itemID]; ball != nil {
			ball.remaining = 0
			ball.resetting = false
		}
	}
	service.mutex.Unlock()
}

// scoreFootballGoal detects directional goal entry and updates its counter.
func (service *Service) scoreFootballGoal(ctx context.Context, active *roomlive.Room, point grid.Point, direction uint8, kickerID int64) (bool, error) {
	for _, goal := range active.FurnitureAt(point) {
		kind := goal.Definition.InteractionType
		if !strings.HasPrefix(kind, "football_goal_") || !football.GoalScores(direction, uint8(goal.Rotation)) {
			continue
		}
		color := strings.TrimPrefix(kind, "football_goal_")
		for _, counter := range active.FurnitureByInteraction("football_counter_" + color) {
			previous, _ := strconv.Atoi(counter.ExtraData)
			next := (previous + 1) % 100
			service.coordinator.NotifyScore(active.ID(), kickerID, int64(previous), int64(next))
			if err := service.projectState(ctx, active, counter.ID, next); err != nil {
				return false, err
			}
		}
		service.progress(ctx, kickerID, "game.football.goal", 1)
		service.metrics.footballGoals.Add(1)
		return true, nil
	}
	return false, nil
}
