package essential

import (
	"context"
	"sort"
	"time"

	roomrealm "github.com/niflaot/pixels/internal/realm/room"
	leavecmd "github.com/niflaot/pixels/internal/realm/room/access/commands/leave"
	"github.com/niflaot/pixels/internal/realm/room/runtime/broadcast"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
	"github.com/niflaot/pixels/networking/codec"
	outalert "github.com/niflaot/pixels/networking/outbound/session/alert"
	"github.com/niflaot/pixels/pkg/i18n"
)

const (
	// cannonFireDelay stores the delay between lighting and firing.
	cannonFireDelay = 750 * time.Millisecond
	// cannonCooldown stores the minimum interval between shots.
	cannonCooldown = 2 * time.Second
)

// useCannon starts one environmental kick shot.
func (service *Service) useCannon(ctx context.Context, request Request) error {
	unit, found := request.Room.Unit(request.PlayerID)
	if !found {
		return nil
	}
	if unit.Moving {
		service.awaitCannon(context.WithoutCancel(ctx), request)
		return nil
	}
	if !adjacentToItem(unit.Position.Point, request.Item) {
		candidates := activatorPoints(request.Item)
		sort.Slice(candidates, func(left int, right int) bool {
			return distance(unit.Position.Point, candidates[left]) < distance(unit.Position.Point, candidates[right])
		})
		for _, point := range candidates {
			roomPath, err := request.Room.MoveControlled(request.PlayerID, point, worldunit.ControlFurnitureInteraction)
			if err == nil && roomPath.Len() > 0 {
				service.awaitCannon(context.WithoutCancel(ctx), request)
				return nil
			}
		}
		return nil
	}
	now := time.Now()
	if !request.Room.TryLockInteraction(request.Item.ID, now.Add(cannonCooldown)) {
		return nil
	}
	if _, err := request.Room.FaceTo(request.PlayerID, request.Item.Point); err != nil {
		request.Room.UnlockInteraction(request.Item.ID)
		return err
	}
	frozen, err := request.Room.SetUnitControl(request.PlayerID, worldunit.ControlFrozen)
	if err != nil {
		request.Room.UnlockInteraction(request.Item.ID)
		return err
	}
	if err := broadcast.RoomUnitStatus(ctx, service.connections, request.Room, frozen, 0); err != nil {
		request.Room.UnlockInteraction(request.Item.ID)
		return err
	}
	next := "1"
	if request.Item.ExtraData == "1" {
		next = "0"
	}
	if err := service.visual(ctx, request.Room, request.Item.ID, next); err != nil {
		request.Room.UnlockInteraction(request.Item.ID)
		return err
	}
	if err := service.publishUsed(ctx, request); err != nil {
		request.Room.UnlockInteraction(request.Item.ID)
		return err
	}
	async := context.WithoutCancel(ctx)
	request.Room.Schedule(cannonFireDelay, func(time.Time) {
		service.fireCannon(async, request)
	})
	request.Room.Schedule(cannonCooldown, func(time.Time) {
		request.Room.UnlockInteraction(request.Item.ID)
	})

	return nil
}

// awaitCannon preserves a click sent while Nitro finishes its automatic approach.
func (service *Service) awaitCannon(ctx context.Context, request Request) {
	request.Room.ScheduleReplacing(scheduledKey(request.Item.ID, 8), roomlive.DefaultTickInterval, func(time.Time) {
		unit, found := request.Room.Unit(request.PlayerID)
		if !found {
			return
		}
		if unit.Moving {
			service.awaitCannon(ctx, request)
			return
		}
		current, found := request.Room.FurnitureItem(request.Item.ID)
		if found {
			request.Item = current
			_ = service.useCannon(ctx, request)
		}
	})
}

// fireCannon releases the activator and removes vulnerable targets in its line.
func (service *Service) fireCannon(ctx context.Context, request Request) {
	if unit, err := request.Room.ReleaseUnitControl(request.PlayerID); err == nil {
		_ = broadcast.RoomUnitStatus(ctx, service.connections, request.Room, unit, 0)
	}
	current, found := request.Room.FurnitureItem(request.Item.ID)
	if !found || current.Point != request.Item.Point || current.Rotation != request.Item.Rotation {
		return
	}
	leave := leavecmd.Handler{
		Players: service.players, Runtime: service.runtime, Connections: service.connections, Events: service.events,
	}
	notice, noticeErr := service.cannonNotice()
	for _, point := range cannonLine(request.Item.Point, request.Item.Rotation) {
		for _, unit := range request.Room.Units() {
			if unit.Position.Point != point || unit.PlayerID == request.Room.Snapshot().OwnerPlayerID {
				continue
			}
			immune := false
			if service.permissions != nil {
				immune, _ = service.permissions.HasPermission(ctx, unit.PlayerID, roomrealm.Unkickable)
			}
			if immune {
				continue
			}
			if noticeErr == nil {
				_ = leave.ToDesktopThen(ctx, unit.PlayerID, notice)
			} else {
				_ = leave.ToDesktop(ctx, unit.PlayerID)
			}
		}
	}
}

// cannonNotice creates the localized environmental kick notice.
func (service *Service) cannonNotice() (packet codec.Packet, err error) {
	message := "You were fired out of the room by a cannon."
	if service.translations != nil {
		message = service.translations.Default(i18n.Key("room.furniture.cannon.kicked"))
	}

	return outalert.Encode(message)
}

// cannonLine returns up to three valid tiles in front of a cannon.
func cannonLine(origin grid.Point, rotation worldunit.Rotation) []grid.Point {
	points := make([]grid.Point, 0, 3)
	current := origin
	rotation = (rotation + 6) % 8
	for range 3 {
		next, ok := offsetPoint(current, rotation)
		if !ok {
			break
		}
		points = append(points, next)
		current = next
	}

	return points
}
