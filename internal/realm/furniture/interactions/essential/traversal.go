package essential

import (
	"context"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/niflaot/pixels/internal/realm/furniture/access"
	"github.com/niflaot/pixels/internal/realm/room/runtime/broadcast"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	worldfurniture "github.com/niflaot/pixels/internal/realm/room/world/furniture"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
)

// useTraversal routes movement-aware and variable-height interactions.
func (service *Service) useTraversal(ctx context.Context, request Request) error {
	switch request.Item.Definition.InteractionType {
	case "onewaygate":
		return service.useOneWayGate(ctx, request)
	case "switch":
		return service.useSwitch(ctx, request, false)
	case "switch_remote_control":
		return service.useSwitch(ctx, request, true)
	default:
		return service.useMultiheight(ctx, request)
	}
}

// useSwitch toggles immediately or after walking to the nearest activator.
func (service *Service) useSwitch(ctx context.Context, request Request, remote bool) error {
	if remote {
		allowed, err := access.CanManage(ctx, service.permissions, request.Room, request.PlayerID)
		if err != nil {
			return err
		}
		if !allowed {
			return ErrNoRights
		}
		return service.toggleFinal(ctx, request)
	}
	unit, found := request.Room.Unit(request.PlayerID)
	if !found {
		return nil
	}
	if adjacentToItem(unit.Position.Point, request.Item) {
		return service.toggleFinal(ctx, request)
	}
	candidates := activatorPoints(request.Item)
	sort.Slice(candidates, func(left int, right int) bool {
		return distance(unit.Position.Point, candidates[left]) < distance(unit.Position.Point, candidates[right])
	})
	for _, point := range candidates {
		roomPath, err := request.Room.MoveControlled(request.PlayerID, point, worldunit.ControlFurnitureInteraction)
		if err == nil && roomPath.Len() > 0 {
			service.awaitSwitch(context.WithoutCancel(ctx), request, point)
			return nil
		}
	}

	return nil
}

// awaitSwitch polls one controlled walk and toggles after arrival.
func (service *Service) awaitSwitch(ctx context.Context, request Request, target grid.Point) {
	request.Room.ScheduleReplacing(scheduledKey(request.Item.ID, 3), roomlive.DefaultTickInterval, func(time.Time) {
		unit, found := request.Room.Unit(request.PlayerID)
		if !found {
			return
		}
		if unit.Moving {
			service.awaitSwitch(ctx, request, target)
			return
		}
		defer request.Room.ReleaseUnitControl(request.PlayerID)
		if unit.Position.Point != target {
			return
		}
		current, found := request.Room.FurnitureItem(request.Item.ID)
		if found {
			request.Item = current
			_ = service.toggleFinal(ctx, request)
		}
	})
}

// useOneWayGate starts one temporary controlled crossing from its front tile.
func (service *Service) useOneWayGate(ctx context.Context, request Request) error {
	unit, found := request.Room.Unit(request.PlayerID)
	front, frontOK := offsetPoint(request.Item.Point, request.Item.Rotation)
	back, backOK := offsetPoint(request.Item.Point, (request.Item.Rotation+4)%8)
	if !found || !frontOK || !backOK || unit.Position.Point != front || pointOccupied(request.Room, request.Item.Point, request.PlayerID) {
		return nil
	}
	if err := service.visual(ctx, request.Room, request.Item.ID, "1"); err != nil {
		return err
	}
	if err := service.publishUsed(ctx, request); err != nil {
		return err
	}
	if err := request.Room.StepControlledOntoInteraction(request.PlayerID, request.Item.Point, worldunit.ControlFurnitureInteraction); err != nil {
		_ = service.visual(ctx, request.Room, request.Item.ID, "0")
		return nil
	}
	async := context.WithoutCancel(ctx)
	request.Room.ScheduleReplacing(scheduledKey(request.Item.ID, 4), roomlive.DefaultTickInterval, func(time.Time) {
		unit, found := request.Room.Unit(request.PlayerID)
		if !found || unit.Position.Point != request.Item.Point {
			service.resetCrossing(async, request)
			return
		}
		if err := request.Room.StepControlledFromInteraction(request.PlayerID, back, worldunit.ControlFurnitureInteraction); err != nil {
			service.resetCrossing(async, request)
			return
		}
		request.Room.ScheduleReplacing(scheduledKey(request.Item.ID, 4), roomlive.DefaultTickInterval, func(time.Time) {
			service.resetCrossing(async, request)
		})
	})

	return nil
}

// resetCrossing closes a one-way gate and releases unit control.
func (service *Service) resetCrossing(ctx context.Context, request Request) {
	_, _ = request.Room.ReleaseUnitControl(request.PlayerID)
	_ = service.visual(ctx, request.Room, request.Item.ID, "0")
}

// useMultiheight changes one physical furniture height.
func (service *Service) useMultiheight(ctx context.Context, request Request) error {
	allowed, err := access.CanManage(ctx, service.permissions, request.Room, request.PlayerID)
	if err != nil {
		return err
	}
	if !allowed {
		return ErrNoRights
	}
	heights := strings.Split(request.Item.Definition.Multiheight, ";")
	if len(heights) < 2 || !request.Room.CanChangeFurnitureHeight(request.Item.ID) {
		return nil
	}
	current, err := strconv.Atoi(request.Item.ExtraData)
	if err != nil || current < 0 || current >= len(heights) {
		current = 0
	}
	next := strconv.Itoa((current + 1) % len(heights))
	if err := service.settle(ctx, request.Room, request.Item.ID, request.Item.ExtraData, next, true); err != nil {
		return err
	}
	if err := service.publishUsed(ctx, request); err != nil {
		return err
	}
	units := request.Room.ResettleFurnitureUnits(request.Item.ID)
	if err := broadcast.RoomUnitStatuses(ctx, service.connections, request.Room, units, 0); err != nil {
		return err
	}
	points := worldfurniture.Footprint(request.Item.Point, request.Item.Definition.Width, request.Item.Definition.Length, request.Item.Rotation)

	return broadcast.RoomHeightMapUpdate(ctx, service.connections, request.Room, points, 0)
}

// toggleFinal applies one durable generic toggle exactly once.
func (service *Service) toggleFinal(ctx context.Context, request Request) error {
	modes := request.Item.Definition.InteractionModesCount
	if modes <= 1 {
		return nil
	}
	current, err := strconv.Atoi(request.Item.ExtraData)
	if err != nil || current < 0 || current >= modes {
		current = 0
	}

	if err := service.settle(ctx, request.Room, request.Item.ID, request.Item.ExtraData, strconv.Itoa((current+1)%modes), false); err != nil {
		return err
	}

	return service.publishUsed(ctx, request)
}
