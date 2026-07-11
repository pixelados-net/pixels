package teleport

import (
	"context"
	"strconv"

	teleportcompleted "github.com/niflaot/pixels/internal/realm/furniture/events/teleportcompleted"
	"github.com/niflaot/pixels/internal/realm/room/runtime/broadcast"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	worldfurniture "github.com/niflaot/pixels/internal/realm/room/world/furniture"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
	outupdate "github.com/niflaot/pixels/networking/outbound/room/furniture/update"
	"github.com/niflaot/pixels/pkg/bus"
)

// openSource projects the source opening and controlled unit placement.
func (service *Service) openSource(ctx context.Context, active *roomlive.Room, transit Transit) error {
	unit, err := active.TeleportUnit(transit.PlayerID, transit.Source.Point, opposite(transit.Source.Rotation), true)
	if err != nil {
		return service.fail(ctx, active.ID(), transit, "source_unavailable")
	}
	item, found := active.SetFurnitureExtraData(transit.Source.ID, "1")
	if !found {
		return service.fail(ctx, active.ID(), transit, "source_removed")
	}
	if err := service.broadcastItem(ctx, active, item); err != nil {
		return err
	}

	return broadcast.RoomUnitStatus(ctx, service.connections, active, unit, 0)
}

// cross transfers the unit to the paired item or forwards the client cross-room.
func (service *Service) cross(ctx context.Context, active *roomlive.Room, transit Transit) (bool, error) {
	if source, found := active.SetFurnitureExtraData(transit.Source.ID, "0"); found {
		_ = service.broadcastItem(ctx, active, source)
	}
	if transit.TargetRoomID != active.ID() {
		return true, service.forward(ctx, active, transit)
	}
	target, found := active.FurnitureItem(transit.Target.ID)
	if !found {
		return true, service.fail(ctx, active.ID(), transit, "target_removed")
	}
	unit, err := active.TeleportUnit(transit.PlayerID, target.Point, target.Rotation, true)
	if err != nil {
		return true, service.fail(ctx, active.ID(), transit, "target_unavailable")
	}
	target, _ = active.SetFurnitureExtraData(target.ID, "2")
	if err := service.broadcastItem(ctx, active, target); err != nil {
		return true, err
	}

	return false, broadcast.RoomUnitStatus(ctx, service.connections, active, unit, 0)
}

// finish releases unit control and closes the destination teleport.
func (service *Service) finish(ctx context.Context, active *roomlive.Room, transit Transit) error {
	unit, found := active.Unit(transit.PlayerID)
	if found {
		unit, _ = active.TeleportUnit(transit.PlayerID, unit.Position.Point, unit.BodyRotation, false)
		_ = broadcast.RoomUnitStatus(ctx, service.connections, active, unit, 0)
	}
	if target, found := active.SetFurnitureExtraData(transit.Target.ID, "0"); found {
		_ = service.broadcastItem(ctx, active, target)
	}
	if service.events == nil {
		return nil
	}

	return service.events.Publish(ctx, bus.Event{Name: teleportcompleted.Name, Payload: teleportcompleted.Payload{
		PlayerID: transit.PlayerID, SourceItemID: transit.Source.ID, SourceRoomID: transit.SourceRoomID,
		TargetItemID: transit.Target.ID, TargetRoomID: active.ID(),
	}})
}

// frontPoint resolves the cardinal tile in front of a teleport.
func frontPoint(item worldfurniture.Item) (grid.Point, bool) {
	x, y := int(item.Point.X), int(item.Point.Y)
	switch item.Rotation {
	case worldunit.RotationNorth:
		y--
	case worldunit.RotationEast:
		x++
	case worldunit.RotationSouth:
		y++
	case worldunit.RotationWest:
		x--
	default:
		return grid.Point{}, false
	}

	return grid.NewPoint(x, y)
}

// opposite returns the opposite Habbo cardinal rotation.
func opposite(rotation worldunit.Rotation) worldunit.Rotation {
	return worldunit.Rotation((int(rotation) + 4) % 8)
}

// broadcastItem sends one runtime furniture visual state.
func (service *Service) broadcastItem(ctx context.Context, active *roomlive.Room, item worldfurniture.Item) error {
	packet, err := outupdate.Encode(outupdate.FloorItem{
		ID: item.ID, SpriteID: item.Definition.SpriteID, X: int(item.Point.X), Y: int(item.Point.Y),
		Rotation: int(item.Rotation), Z: strconv.Itoa(int(item.Z)), ExtraData: item.ExtraData, OwnerID: item.OwnerPlayerID,
	})
	if err != nil {
		return err
	}

	return broadcast.RoomPacket(ctx, service.connections, active, packet, 0)
}
