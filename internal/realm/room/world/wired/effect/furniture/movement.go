package furniture

import (
	"context"

	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	worldfurniture "github.com/niflaot/pixels/internal/realm/room/world/furniture"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/configuration"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/effect"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/trigger"
)

// move resolves one valid adjacent destination or rotation.
func (service *Service) move(ctx context.Context, active *roomlive.Room, itemID int64, operation effect.FurnitureOperation, node *configuration.Node, event trigger.Event) (mutation, error) {
	item, found := active.FurnitureItem(itemID)
	if !found {
		return mutation{}, nil
	}
	if operation == effect.FleeActor {
		if unit, collided := adjacentUnit(active, item); collided {
			return collisionMutation(active.ID(), item, unit), nil
		}
	}
	if operation == effect.MoveRotate && movementMode(node) == 0 {
		rotation := worldunit.Rotation((int(item.Rotation) + 2) % 8)
		return service.placeMovement(ctx, active, item, item.Point, rotation, event)
	}
	candidates := movementCandidates(active, item, operation, node, event)
	for _, point := range candidates {
		if point == item.Point {
			continue
		}
		result, err := service.placeMovement(ctx, active, item, point, item.Rotation, event)
		if err != nil {
			return mutation{}, err
		}
		if result.changed || result.collided {
			return result, nil
		}
	}
	return mutation{}, nil
}

// placeMovement detects unit collisions before applying normal authoritative placement.
func (service *Service) placeMovement(ctx context.Context, active *roomlive.Room, item worldfurniture.Item, point grid.Point, rotation worldunit.Rotation, event trigger.Event) (mutation, error) {
	candidate := item
	candidate.Point = point
	candidate.Rotation = rotation
	if unit, collided := active.UnitInFurnitureFootprint(candidate); collided {
		return collisionMutation(active.ID(), candidate, unit), nil
	}
	changed, err := service.place(ctx, active, item, point, rotation, event.PlayerID)

	return mutation{changed: changed}, err
}

// adjacentUnit returns the stable first unit touching a flee target.
func adjacentUnit(active *roomlive.Room, item worldfurniture.Item) (roomlive.UnitSnapshot, bool) {
	for _, unit := range active.Units() {
		if distance(item.Point, unit.Position.Point) <= 1 {
			return unit, true
		}
	}

	return roomlive.UnitSnapshot{}, false
}

// collisionMutation maps a collided room unit into a derived WIRED collision event.
func collisionMutation(roomID int64, item worldfurniture.Item, unit roomlive.UnitSnapshot) mutation {
	actor := trigger.ActorPlayer
	if unit.Kind == worldunit.KindBot {
		actor = trigger.ActorBot
	} else if unit.Kind == worldunit.KindPet {
		actor = trigger.ActorPet
	}

	return mutation{collided: true, collision: trigger.Event{
		Kind: trigger.Collision, RoomID: roomID, ActorKind: actor, ActorID: unit.EntityKey,
		PlayerID: unit.PlayerID, SourceItem: item.ID, SourceSprite: int32(item.Definition.SpriteID),
	}}
}

// place validates, persists, swaps runtime state, and broadcasts a placement.
func (service *Service) place(ctx context.Context, active *roomlive.Room, item worldfurniture.Item, point grid.Point, rotation worldunit.Rotation, actorID int64) (bool, error) {
	if point == item.Point && rotation == item.Rotation {
		return false, nil
	}
	resolved, valid, err := service.resolveWorldItem(ctx, active, item.ID, point, rotation)
	if err != nil || !valid {
		return false, nil
	}
	durable, found, err := service.furniture.FindItemByID(ctx, item.ID)
	if err != nil || !found {
		return false, err
	}
	if actorID <= 0 {
		actorID = durable.OwnerPlayerID
	}
	_, err = service.furniture.Move(ctx, furnitureservice.MoveParams{ItemID: item.ID, ActorPlayerID: actorID, RoomID: active.ID(), Placement: furnituremodel.Placement{X: int(point.X), Y: int(point.Y), Z: resolved.Z.Units(), Rotation: durableRotation(rotation)}})
	if err != nil {
		return false, err
	}
	if _, err = active.ReloadFurniture(item.ID, &resolved); err != nil {
		return false, err
	}
	return true, service.broadcast(ctx, active, resolved)
}

// movementCandidates returns stable adjacent candidates in preference order.
func movementCandidates(active *roomlive.Room, item worldfurniture.Item, operation effect.FurnitureOperation, node *configuration.Node, event trigger.Event) []grid.Point {
	points := cardinal(item.Point)
	if operation == effect.MoveDirection || operation == effect.MoveFurnitureTo || operation == effect.MoveRotate {
		direction := movementMode(node)
		if point, ok := grid.PointInFront(item.Point, uint8(direction)); ok {
			return []grid.Point{point}
		}
		return nil
	}
	unit, found := active.UnitMotion(event.ActorID)
	if !found {
		return points
	}
	target := unit.Position.Point
	away := operation == effect.FleeActor
	sortByDistance(points, target, away)
	return points
}

// cardinal returns valid cardinal neighbors.
func cardinal(point grid.Point) []grid.Point {
	result := make([]grid.Point, 0, 4)
	for _, rotation := range []uint8{0, 2, 4, 6} {
		if candidate, valid := grid.PointInFront(point, rotation); valid {
			result = append(result, candidate)
		}
	}
	return result
}

// sortByDistance stably orders points toward or away from a target.
func sortByDistance(points []grid.Point, target grid.Point, away bool) {
	for left := 0; left < len(points); left++ {
		for right := left + 1; right < len(points); right++ {
			leftDistance := distance(points[left], target)
			rightDistance := distance(points[right], target)
			if (!away && rightDistance < leftDistance) || (away && rightDistance > leftDistance) {
				points[left], points[right] = points[right], points[left]
			}
		}
	}
}

// distance returns Manhattan tile distance.
func distance(left grid.Point, right grid.Point) int {
	dx := int(left.X) - int(right.X)
	dy := int(left.Y) - int(right.Y)
	if dx < 0 {
		dx = -dx
	}
	if dy < 0 {
		dy = -dy
	}
	return dx + dy
}

// movementMode returns the first configured direction or zero.
func movementMode(node *configuration.Node) int32 {
	if len(node.Parameters.Values) == 0 {
		return 0
	}
	return node.Parameters.Values[0]
}
