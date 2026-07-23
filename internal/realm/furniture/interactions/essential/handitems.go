package essential

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/niflaot/pixels/internal/realm/room/runtime/broadcast"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	worldfurniture "github.com/niflaot/pixels/internal/realm/room/world/furniture"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
	outhanditem "github.com/niflaot/pixels/networking/outbound/room/entities/handitem"
)

// useHandItem routes vending and direct hand-item interactions.
func (service *Service) useHandItem(ctx context.Context, request Request) error {
	if request.Item.Definition.InteractionType == "handitem" {
		unit, found := request.Room.Unit(request.PlayerID)
		if !found || (unit.Position.Point != request.Item.Point && !adjacentToItem(unit.Position.Point, request.Item)) {
			return nil
		}

		if err := service.giveHandItem(ctx, request.Room, request.PlayerID, request.Item, true); err != nil {
			return err
		}

		return service.publishUsed(ctx, request)
	}
	unit, found := request.Room.Unit(request.PlayerID)
	if !found {
		return nil
	}
	activators := vendingActivators(request.Item)
	if containsPoint(activators, unit.Position.Point) {
		return service.activateVending(ctx, request)
	}
	for _, point := range activators {
		roomPath, err := request.Room.MoveControlled(request.PlayerID, point, worldunit.ControlFurnitureInteraction)
		if err == nil && roomPath.Len() > 0 {
			service.awaitVending(context.WithoutCancel(ctx), request, point)
			return nil
		}
	}

	return nil
}

// awaitVending activates after one controlled walk settles.
func (service *Service) awaitVending(ctx context.Context, request Request, target grid.Point) {
	request.Room.ScheduleReplacing(scheduledKey(request.Item.ID, 5), roomlive.DefaultTickInterval, func(time.Time) {
		unit, found := request.Room.Unit(request.PlayerID)
		if !found {
			return
		}
		if unit.Moving {
			service.awaitVending(ctx, request, target)
			return
		}
		defer request.Room.ReleaseUnitControl(request.PlayerID)
		if unit.Position.Point == target {
			_ = service.activateVending(ctx, request)
		}
	})
}

// activateVending animates one machine and delivers after its delay.
func (service *Service) activateVending(ctx context.Context, request Request) error {
	if request.Item.ExtraData == "1" {
		return nil
	}
	_, _ = request.Room.FaceTo(request.PlayerID, request.Item.Point)
	if err := service.visual(ctx, request.Room, request.Item.ID, "1"); err != nil {
		return err
	}
	if err := service.publishUsed(ctx, request); err != nil {
		return err
	}
	async := context.WithoutCancel(ctx)
	request.Room.ScheduleReplacing(scheduledKey(request.Item.ID, 6), 1500*time.Millisecond, func(time.Time) {
		current, found := request.Room.FurnitureItem(request.Item.ID)
		if !found || current.Definition.InteractionType != request.Item.Definition.InteractionType {
			return
		}
		_ = service.giveHandItem(async, request.Room, request.PlayerID, request.Item, false)
		request.Room.ScheduleReplacing(scheduledKey(request.Item.ID, 6), 500*time.Millisecond, func(time.Time) {
			_ = service.visual(async, request.Room, request.Item.ID, "0")
		})
	})

	return nil
}

// giveHandItem selects and broadcasts one carried item.
func (service *Service) giveHandItem(ctx context.Context, active *roomlive.Room, playerID int64, item worldfurniture.Item, animate bool) error {
	items := handItemIDs(item.Definition.CustomParams)
	if len(items) == 0 {
		return nil
	}
	selected := items[0]
	if len(items) > 1 {
		selected = items[service.random.IntN(len(items))]
	}
	unit, err := active.SetHandItem(playerID, selected)
	if err != nil {
		return err
	}
	packet, err := outhanditem.Encode(unit.UnitID, selected)
	if err != nil {
		return err
	}
	if err := broadcast.RoomPacket(ctx, service.connections, active, packet, 0); err != nil {
		return err
	}
	if animate && item.Definition.InteractionModesCount > 1 {
		if err := service.visual(ctx, active, item.ID, "1"); err != nil {
			return err
		}
		async := context.WithoutCancel(ctx)
		active.ScheduleReplacing(scheduledKey(item.ID, 7), 500*time.Millisecond, func(time.Time) {
			_ = service.visual(async, active, item.ID, "0")
		})
	}

	return nil
}

// handItemIDs parses definition vending ids.
func handItemIDs(value string) []int32 {
	parts := strings.Split(value, ",")
	items := make([]int32, 0, len(parts))
	for _, part := range parts {
		itemID, err := strconv.ParseInt(strings.TrimSpace(part), 10, 32)
		if err == nil && itemID > 0 {
			items = append(items, int32(itemID))
		}
	}

	return items
}

// vendingActivators resolves normal and no-sides activation tiles.
func vendingActivators(item worldfurniture.Item) []grid.Point {
	if item.Definition.InteractionType == "vendingmachine_no_sides" {
		return activatorPoints(item)
	}
	points := []grid.Point{item.Point}
	if front, ok := offsetPoint(item.Point, item.Rotation); ok {
		points = append(points, front)
	}

	return points
}

// containsPoint reports whether a point belongs to one small list.
func containsPoint(points []grid.Point, target grid.Point) bool {
	for _, point := range points {
		if point == target {
			return true
		}
	}

	return false
}
