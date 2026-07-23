// Package furnitureindex builds and queries immutable room furniture indexes.
package furnitureindex

import (
	"fmt"

	worldfurniture "github.com/niflaot/pixels/internal/realm/room/world/furniture"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	"github.com/niflaot/pixels/internal/realm/room/world/surface"
)

// Snapshot stores the indexes built from one furniture generation.
type Snapshot struct {
	// Fixtures stores every physical furniture fixture.
	Fixtures []surface.Fixture
	// Items stores furniture by durable id.
	Items map[int64]worldfurniture.Item
	// Tiles stores furniture ids by footprint tile.
	Tiles map[grid.Point][]int64
	// Types stores furniture ids by interaction type.
	Types map[string][]int64
	// Interactions stores interactive furniture ids by footprint tile.
	Interactions map[grid.Point]int64
}

// Build creates furniture indexes for one loaded room generation.
func Build(items []worldfurniture.Item) (Snapshot, error) {
	result := Snapshot{
		Fixtures: make([]surface.Fixture, 0, len(items)), Items: make(map[int64]worldfurniture.Item, len(items)),
		Tiles: make(map[grid.Point][]int64), Types: make(map[string][]int64),
	}
	for _, item := range items {
		itemFixtures, err := worldfurniture.Fixtures(item)
		if err != nil {
			return Snapshot{}, fmt.Errorf("build fixtures for furniture item %d: %w", item.ID, err)
		}
		result.Fixtures = append(result.Fixtures, itemFixtures...)
		result.Items[item.ID] = item
		for _, point := range worldfurniture.Footprint(item.Point, item.Definition.Width, item.Definition.Length, item.Rotation) {
			result.Tiles[point] = append(result.Tiles[point], item.ID)
		}
		if item.Definition.InteractionType != "" {
			result.Types[item.Definition.InteractionType] = append(result.Types[item.Definition.InteractionType], item.ID)
		}
		if interactive(item) {
			if result.Interactions == nil {
				result.Interactions = make(map[grid.Point]int64)
			}
			for _, point := range worldfurniture.Footprint(item.Point, item.Definition.Width, item.Definition.Length, item.Rotation) {
				result.Interactions[point] = item.ID
			}
		}
	}

	return result, nil
}

// ResolveSlotGoal maps a furniture footprint target onto a usable slot.
func ResolveSlotGoal(items map[int64]worldfurniture.Item, goal grid.Point) grid.Point {
	for _, item := range items {
		slots := worldfurniture.Slots(item)
		if len(slots) == 0 || !contains(item, goal) {
			continue
		}
		for _, slot := range slots {
			if slot.Point == goal {
				return goal
			}
		}
		for _, slot := range slots {
			if slot.Point.X == goal.X || slot.Point.Y == goal.Y {
				return slot.Point
			}
		}
		return slots[0].Point
	}

	return goal
}

// ByInteraction returns stable furniture snapshots for one indexed interaction type.
func ByInteraction(ids []int64, items map[int64]worldfurniture.Item) []worldfurniture.Item {
	if len(ids) == 0 {
		return nil
	}
	selected := make([]worldfurniture.Item, 0, len(ids))
	for _, id := range ids {
		if item, found := items[id]; found {
			selected = append(selected, item)
		}
	}
	return selected
}

// Nearest returns the closest indexed item without allocating.
func Nearest(ids []int64, items map[int64]worldfurniture.Item, origin grid.Point, radius int) (worldfurniture.Item, bool) {
	selected, selectedDistance, found := worldfurniture.Item{}, 0, false
	for _, id := range ids {
		item, exists := items[id]
		if !exists {
			continue
		}
		distance := distance(origin, item.Point)
		if radius >= 0 && distance > radius {
			continue
		}
		if !found || distance < selectedDistance || distance == selectedDistance && item.ID < selected.ID {
			selected, selectedDistance, found = item, distance, true
		}
	}
	return selected, found
}

// contains reports whether a point belongs to an item's footprint.
func contains(item worldfurniture.Item, point grid.Point) bool {
	for _, current := range worldfurniture.Footprint(item.Point, item.Definition.Width, item.Definition.Length, item.Rotation) {
		if current == point {
			return true
		}
	}
	return false
}

// distance returns the Manhattan distance between two room tiles.
func distance(first grid.Point, second grid.Point) int {
	dx, dy := abs(int(first.X)-int(second.X)), abs(int(first.Y)-int(second.Y))
	return dx + dy
}

// abs returns a non-negative integer magnitude.
func abs(value int) int {
	if value < 0 {
		return -value
	}
	return value
}

// interactive reports whether an item participates in movement interaction events.
func interactive(item worldfurniture.Item) bool {
	return item.Definition.InteractionType != "" &&
		(item.Definition.InteractionType != "default" || item.Definition.InteractionModesCount > 1)
}
