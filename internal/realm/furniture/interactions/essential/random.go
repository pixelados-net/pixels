package essential

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/niflaot/pixels/internal/realm/furniture/access"
	randomresolved "github.com/niflaot/pixels/internal/realm/furniture/events/randomresolved"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	roomtask "github.com/niflaot/pixels/internal/realm/room/runtime/live/task"
	worldfurniture "github.com/niflaot/pixels/internal/realm/room/world/furniture"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
	"github.com/niflaot/pixels/pkg/bus"
	"go.uber.org/zap"
)

const (
	// randomRollingState stores Nitro's rolling sentinel.
	randomRollingState = "-1"
)

// useRandom starts one delayed random-state interaction.
func (service *Service) useRandom(ctx context.Context, request Request) error {
	if request.Item.ExtraData == randomRollingState {
		return nil
	}
	modes, delay, err := service.randomPolicy(ctx, request)
	if err != nil || modes <= 0 {
		return err
	}
	if request.Item.Definition.InteractionType == "dice" {
		unit, found := request.Room.Unit(request.PlayerID)
		if !found || !adjacent(unit.Position.Point, request.Item.Point) {
			return nil
		}
	}
	if err := service.visual(ctx, request.Room, request.Item.ID, randomRollingState); err != nil {
		return err
	}
	if err := service.publishUsed(ctx, request); err != nil {
		return err
	}
	async := context.WithoutCancel(ctx)
	key := scheduledKey(request.Item.ID, 1)
	request.Room.ScheduleReplacing(key, delay, func(time.Time) {
		result := service.random.IntN(modes) + 1
		value := strconv.Itoa(result)
		if err := service.settle(async, request.Room, request.Item.ID, request.Item.ExtraData, value, false); err != nil {
			if service.log != nil {
				service.log.Error("furniture random state resolution failed", zap.Error(err), zap.Int64("item_id", request.Item.ID), zap.Int64("room_id", request.Room.ID()))
			}
			return
		}
		if service.events != nil {
			_ = service.events.Publish(async, bus.Event{Name: randomresolved.Name, Payload: randomresolved.Payload{
				PlayerID: request.PlayerID, ItemID: request.Item.ID, RoomID: request.Room.ID(), Result: result,
			}})
		}
	})

	return nil
}

// CloseDice resets one settled dice value through Nitro's dedicated close request.
func (service *Service) CloseDice(ctx context.Context, request Request) error {
	if request.Room == nil || request.Item.Definition.InteractionType != "dice" || request.Item.ExtraData == randomRollingState || request.Item.ExtraData == "0" {
		return nil
	}

	return service.settle(ctx, request.Room, request.Item.ID, request.Item.ExtraData, "0", false)
}

// randomPolicy resolves state count, delay, and authorization.
func (service *Service) randomPolicy(ctx context.Context, request Request) (int, time.Duration, error) {
	switch request.Item.Definition.InteractionType {
	case "dice":
		return request.Item.Definition.InteractionModesCount, 1500 * time.Millisecond, nil
	case "colorwheel":
		allowed, err := access.CanManage(ctx, service.permissions, request.Room, request.PlayerID)
		if err != nil || !allowed {
			if err == nil {
				err = ErrNoRights
			}
			return 0, 0, err
		}
		return request.Item.Definition.InteractionModesCount, 3 * time.Second, nil
	default:
		states, delay := parseRandomParams(request.Item.Definition.CustomParams)
		return states, delay, nil
	}
}

// parseRandomParams parses Arcturus-compatible states and delay parameters.
func parseRandomParams(value string) (int, time.Duration) {
	states := 0
	delay := time.Duration(0)
	for _, pair := range strings.Split(value, ",") {
		key, raw, found := strings.Cut(pair, "=")
		if !found {
			continue
		}
		number, err := strconv.Atoi(strings.TrimSpace(raw))
		if err != nil || number < 0 {
			continue
		}
		switch strings.TrimSpace(key) {
		case "states":
			states = number
		case "delay":
			delay = time.Duration(number) * time.Millisecond
		}
	}

	return states, delay
}

// adjacent reports whether points share one edge or corner.
func adjacent(first grid.Point, second grid.Point) bool {
	dx := abs(int(first.X) - int(second.X))
	dy := abs(int(first.Y) - int(second.Y))

	return (dx != 0 || dy != 0) && dx <= 1 && dy <= 1
}

// scheduledKey creates a stable non-zero replacement key.
func scheduledKey(itemID int64, kind uint8) roomtask.Key {
	return roomtask.Key(uint64(itemID)<<8 | uint64(kind) | 1)
}

// abs returns an integer magnitude.
func abs(value int) int {
	if value < 0 {
		return -value
	}

	return value
}

// offsetPoint returns the tile in one Habbo rotation direction.
func offsetPoint(point grid.Point, rotation worldunit.Rotation) (grid.Point, bool) {
	offsets := [8][2]int{{0, -1}, {1, -1}, {1, 0}, {1, 1}, {0, 1}, {-1, 1}, {-1, 0}, {-1, -1}}
	offset := offsets[int(rotation)%len(offsets)]

	return grid.NewPoint(int(point.X)+offset[0], int(point.Y)+offset[1])
}

// pointOccupied reports whether another unit occupies a tile.
func pointOccupied(active interface {
	Units() []roomlive.UnitSnapshot
}, point grid.Point, excludedPlayerID int64) bool {
	for _, unit := range active.Units() {
		if unit.PlayerID != excludedPlayerID && unit.Position.Point == point {
			return true
		}
	}

	return false
}

// adjacentToItem reports whether a point borders any footprint tile.
func adjacentToItem(point grid.Point, item worldfurniture.Item) bool {
	for _, footprintPoint := range worldfurniture.Footprint(item.Point, item.Definition.Width, item.Definition.Length, item.Rotation) {
		if adjacent(point, footprintPoint) {
			return true
		}
	}

	return false
}

// activatorPoints returns valid perimeter points around an item.
func activatorPoints(item worldfurniture.Item) []grid.Point {
	width, length := worldfurniture.Dimensions(item.Definition.Width, item.Definition.Length, item.Rotation)
	points := make([]grid.Point, 0, 2*(width+length)+4)
	for y := -1; y <= length; y++ {
		for x := -1; x <= width; x++ {
			if x >= 0 && x < width && y >= 0 && y < length {
				continue
			}
			point, ok := grid.NewPoint(int(item.Point.X)+x, int(item.Point.Y)+y)
			if !ok {
				continue
			}
			points = append(points, point)
		}
	}

	return points
}

// distance returns squared tile distance.
func distance(first grid.Point, second grid.Point) int {
	dx := int(first.X) - int(second.X)
	dy := int(first.Y) - int(second.Y)

	return dx*dx + dy*dy
}
