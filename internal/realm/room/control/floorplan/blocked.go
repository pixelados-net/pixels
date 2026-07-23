package floorplan

import (
	"sort"

	worldfurniture "github.com/niflaot/pixels/internal/realm/room/world/furniture"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
)

// BlockedItem stores one furniture item affected by geometry changes.
type BlockedItem struct {
	// Item stores the affected runtime furniture snapshot.
	Item worldfurniture.Item
}

// OccupiedTiles returns unique furniture footprint tiles in stable row-major order.
func OccupiedTiles(items []worldfurniture.Item) []grid.Point {
	seen := make(map[grid.Point]struct{}, len(items))
	for _, item := range items {
		forEachFootprint(item, func(point grid.Point) bool {
			seen[point] = struct{}{}
			return true
		})
	}
	points := make([]grid.Point, 0, len(seen))
	for point := range seen {
		points = append(points, point)
	}
	sortPoints(points)

	return points
}

// BlockedItems returns furniture whose supporting geometry changes.
func BlockedItems(previous grid.Grid, next grid.Grid, items []worldfurniture.Item) []BlockedItem {
	blocked := make([]BlockedItem, 0, len(items))
	for _, item := range items {
		if footprintChanged(previous, next, item) {
			blocked = append(blocked, BlockedItem{Item: item})
		}
	}
	sort.Slice(blocked, func(left int, right int) bool {
		return blocked[left].Item.ID < blocked[right].Item.ID
	})

	return blocked
}

// footprintChanged reports whether any supporting base tile changed or disappeared.
func footprintChanged(previous grid.Grid, next grid.Grid, item worldfurniture.Item) bool {
	changed := false
	forEachFootprint(item, func(point grid.Point) bool {
		oldHeight, oldValid := previous.HeightAt(point)
		newHeight, newValid := next.HeightAt(point)
		if oldValid != newValid || oldHeight != newHeight {
			changed = true
			return false
		}
		return true
	})

	return changed
}

// forEachFootprint visits one placed item's occupied tiles until visit returns false.
func forEachFootprint(item worldfurniture.Item, visit func(grid.Point) bool) {
	width, length := worldfurniture.Dimensions(item.Definition.Width, item.Definition.Length, item.Rotation)
	for dy := 0; dy < length; dy++ {
		for dx := 0; dx < width; dx++ {
			point, ok := grid.NewPoint(int(item.Point.X)+dx, int(item.Point.Y)+dy)
			if ok && !visit(point) {
				return
			}
		}
	}
}

// sortPoints orders points by row and then column.
func sortPoints(points []grid.Point) {
	sort.Slice(points, func(left int, right int) bool {
		if points[left].Y == points[right].Y {
			return points[left].X < points[right].X
		}

		return points[left].Y < points[right].Y
	})
}
