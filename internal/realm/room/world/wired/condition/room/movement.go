package room

import (
	"sort"

	worldfurniture "github.com/niflaot/pixels/internal/realm/room/world/furniture"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/configuration"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/record"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/trigger"
)

// ValidMoves simulates cumulative movement destinations against the live geometry.
func (view *View) ValidMoves(effects []*configuration.Node, event trigger.Event) (bool, error) {
	shadow := make(map[int64]worldfurniture.Item)
	for _, node := range effects {
		if !movementEffect(node.Descriptor.Key) {
			continue
		}
		for _, target := range node.Targets {
			item, found := shadow[target.ItemID]
			if !found {
				item, found = view.active.FurnitureItem(target.ItemID)
			}
			if !found {
				return false, nil
			}
			candidate, rotation, valid := view.simulatedDestination(item, target.Snapshot, node, event)
			if !valid || !view.validDestination(item, candidate, rotation) {
				return false, nil
			}
			item.Point, item.Rotation = candidate, rotation
			shadow[item.ID] = item
		}
	}
	return true, nil
}

// simulatedDestination returns the first deterministic destination production would attempt.
func (view *View) simulatedDestination(item worldfurniture.Item, snapshot record.Snapshot, node *configuration.Node, event trigger.Event) (grid.Point, worldunit.Rotation, bool) {
	key := node.Descriptor.Key
	if key == "wf_act_match_to_sshot" {
		point, valid := grid.NewPoint(snapshot.X, snapshot.Y)
		if !snapshot.Present {
			return grid.Point{}, item.Rotation, false
		}
		return point, worldunit.Rotation(snapshot.Rotation), valid
	}
	mode := firstValue(node.Parameters.Values)
	if key == "wf_act_move_rotate" && mode == 0 {
		return item.Point, worldunit.Rotation((int(item.Rotation) + 2) % 8), true
	}
	if key == "wf_act_chase" || key == "wf_act_flee" {
		unit, found := view.active.UnitMotion(event.ActorID)
		if !found {
			return grid.Point{}, item.Rotation, false
		}
		points := adjacent(item.Point)
		away := key == "wf_act_flee"
		sort.SliceStable(points, func(left int, right int) bool {
			leftDistance, rightDistance := tileDistance(points[left], unit.Position.Point), tileDistance(points[right], unit.Position.Point)
			if away {
				return leftDistance > rightDistance
			}
			return leftDistance < rightDistance
		})
		if len(points) == 0 {
			return grid.Point{}, item.Rotation, false
		}
		return points[0], item.Rotation, true
	}
	point, valid := grid.PointInFront(item.Point, uint8(mode))
	return point, item.Rotation, valid
}

// validDestination rejects invalid layout tiles and room-unit collisions for a full footprint.
func (view *View) validDestination(item worldfurniture.Item, point grid.Point, rotation worldunit.Rotation) bool {
	for _, tile := range worldfurniture.Footprint(point, item.Definition.Width, item.Definition.Length, rotation) {
		column, err := view.active.SurfaceColumn(tile)
		if err != nil {
			return false
		}
		section, found := column.TopSection()
		if !found || !section.Stacking() {
			return false
		}
		for _, unit := range view.active.Units() {
			if unit.Position.Point == tile {
				return false
			}
		}
	}
	return true
}

// movementEffect reports compatibility-simulated furniture operations.
func movementEffect(key string) bool {
	switch key {
	case "wf_act_match_to_sshot", "wf_act_move_rotate", "wf_act_chase", "wf_act_flee", "wf_act_move_to_dir", "wf_act_move_furni_to":
		return true
	default:
		return false
	}
}

// adjacent returns cardinal room points in stable rotation order.
func adjacent(point grid.Point) []grid.Point {
	result := make([]grid.Point, 0, 4)
	for _, rotation := range []uint8{0, 2, 4, 6} {
		if candidate, valid := grid.PointInFront(point, rotation); valid {
			result = append(result, candidate)
		}
	}
	return result
}

// tileDistance returns Manhattan distance.
func tileDistance(left grid.Point, right grid.Point) int {
	dx, dy := int(left.X)-int(right.X), int(left.Y)-int(right.Y)
	if dx < 0 {
		dx = -dx
	}
	if dy < 0 {
		dy = -dy
	}
	return dx + dy
}

// firstValue returns the first integer setting.
func firstValue(values []int32) int32 {
	if len(values) == 0 {
		return 0
	}
	return values[0]
}
