package games

import (
	"context"
	"strconv"
	"strings"
	"time"

	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	roommoved "github.com/niflaot/pixels/internal/realm/room/world/events/moved"
	worldfurniture "github.com/niflaot/pixels/internal/realm/room/world/furniture"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	"github.com/niflaot/pixels/pkg/bus"
)

// footballBall stores one room-tick movement queue.
type footballBall struct {
	// direction stores one of eight Habbo rotations.
	direction uint8
	// remaining stores decelerating movement steps.
	remaining int
	// kickerID identifies the most recent player kick.
	kickerID int64
	// resetting prevents a scored ball from being kicked before it returns home.
	resetting bool
}

// UseRequest describes one game-owned specialized furniture use.
type UseRequest struct {
	// PlayerID identifies the acting player.
	PlayerID int64
	// Room stores the active room.
	Room *roomlive.Room
	// Item stores the targeted furniture.
	Item worldfurniture.Item
	// State stores Nitro's start-stop or increase-time surface.
	State int32
	// blockID preserves an approached Freeze block across the room queue.
	blockID int64
}

// UseFurniture handles game timers, balls, blocks, poles, and scoreboards.
func (service *Service) UseFurniture(ctx context.Context, request UseRequest) (bool, error) {
	if !service.config.Enabled {
		return false, nil
	}
	kind := request.Item.Definition.InteractionType
	switch {
	case kind == "game_timer":
		if !request.Room.CanManageFurniture(request.PlayerID) {
			return true, nil
		}
		if request.State == 2 {
			return true, service.increaseTimer(ctx, request)
		}
		return true, service.toggleTimer(ctx, request)
	case strings.HasPrefix(kind, "football_counter_"):
		if !request.Room.CanManageFurniture(request.PlayerID) {
			return true, nil
		}
		value, _ := strconv.Atoi(request.Item.ExtraData)
		next := footballCounterNext(value, request.State)
		service.coordinator.NotifyScore(request.Room.ID(), request.PlayerID, int64(value), int64(next))
		return true, service.projectState(ctx, request.Room, request.Item.ID, next)
	case kind == "freeze_block" || kind == "freeze_tile":
		return true, service.throwFreeze(ctx, request)
	case kind == "football":
		return true, service.kickFootball(request)
	case strings.HasSuffix(kind, "_pole"):
		return true, nil
	default:
		return false, nil
	}
}

// footballCounterNext maps Nitro use modes to increment, decrement, and reset.
func footballCounterNext(value int, mode int32) int {
	if mode == 2 {
		return 0
	}
	if mode == 1 {
		return (value + 99) % 100
	}
	return (value + 1) % 100
}

// UnitMoved starts a football kick and transfers Tag when adjacent.
func (service *Service) UnitMoved(ctx context.Context, event bus.Event) error {
	payload, ok := event.Payload.(roommoved.Payload)
	if !ok || payload.PlayerID <= 0 || !service.config.Enabled {
		return nil
	}
	active, found := service.rooms.Find(payload.RoomID)
	if !found {
		return nil
	}
	service.transferTag(active, payload.PlayerID)
	snapshot, found := service.wired.Snapshot(payload.RoomID)
	if !found || !snapshot.Running {
		return nil
	}
	for _, item := range active.FurnitureAt(payload.Current.Point) {
		if item.Definition.InteractionType != "football" && item.Definition.InteractionType != "battlebanzai_puck" {
			continue
		}
		direction := uint8(rotationBetween(payload.Previous.Point, payload.Current.Point))
		service.queueFootball(active, item.ID, direction, payload.PlayerID)
		break
	}
	return nil
}

// cycleFootball advances each active ball one step.
func (service *Service) cycleFootball(ctx context.Context, active *roomlive.Room) error {
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
	ids := make([]int64, 0, len(state.footballs))
	for id, ball := range state.footballs {
		if ball.remaining > 0 {
			ids = append(ids, id)
		}
	}
	service.mutex.Unlock()
	for _, id := range ids {
		if err := service.stepFootball(ctx, active, id); err != nil {
			return err
		}
	}
	return nil
}

// stepFootball resolves collision, movement, rolling projection, and goals.
func (service *Service) stepFootball(ctx context.Context, active *roomlive.Room, itemID int64) error {
	item, found := active.FurnitureItem(itemID)
	if !found || service.furniture == nil {
		return nil
	}
	service.mutex.Lock()
	state := service.states[active.ID()]
	ball := state.footballs[itemID]
	if ball == nil || ball.remaining <= 0 {
		service.mutex.Unlock()
		return nil
	}
	direction, kickerID := ball.direction, ball.kickerID
	target, direction, valid := service.resolveFootballMove(active, item, direction)
	if !valid {
		ball.remaining = 0
		service.mutex.Unlock()
		return nil
	}
	ball.direction = direction
	ball.remaining--
	service.mutex.Unlock()
	if err := service.moveFootball(ctx, active, item, target); err != nil {
		return err
	}
	if item.Definition.InteractionType == "battlebanzai_puck" {
		for _, tile := range active.FurnitureAt(target) {
			if tile.Definition.InteractionType == "battlebanzai_tile" {
				return service.stepBanzai(ctx, active, kickerID, tile.ID)
			}
		}
		return nil
	}
	scored, err := service.scoreFootballGoal(ctx, active, target, direction, kickerID)
	if err != nil || !scored {
		return err
	}
	service.mutex.Lock()
	if state := service.states[active.ID()]; state != nil && state.footballs[item.ID] != nil {
		state.footballs[item.ID].remaining = 0
		state.footballs[item.ID].resetting = true
	}
	service.mutex.Unlock()
	active.Schedule(750*time.Millisecond, func(time.Time) {
		_ = service.resetFootball(context.Background(), active, item.ID)
	})
	return nil
}

// rotationBetween returns one Habbo direction between adjacent points.
func rotationBetween(from grid.Point, to grid.Point) uint8 {
	dx, dy := int(to.X)-int(from.X), int(to.Y)-int(from.Y)
	directions := [][3]int{{0, -1, 0}, {1, -1, 1}, {1, 0, 2}, {1, 1, 3}, {0, 1, 4}, {-1, 1, 5}, {-1, 0, 6}, {-1, -1, 7}}
	for _, direction := range directions {
		if dx == direction[0] && dy == direction[1] {
			return uint8(direction[2])
		}
	}
	return 0
}
