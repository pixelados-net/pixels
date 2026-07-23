package projection

import (
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	outheightmapupdate "github.com/niflaot/pixels/networking/outbound/room/heightmapupdate"
)

const (
	// stackingBlockedBit marks a tile where new items cannot stack, matching Nitro's wire format.
	stackingBlockedBit = 0x4000

	// invalidTileValue marks a tile outside the room, matching Nitro's wire format.
	invalidTileValue = -1
)

// HeightMapTiles maps resolved room tile heights into ROOM_HEIGHT_MAP wire values.
func HeightMapTiles(tiles []roomlive.TileHeight) []int16 {
	values := make([]int16, 0, len(tiles))
	for _, tile := range tiles {
		values = append(values, heightMapValue(tile))
	}

	return values
}

// HeightMapUpdateTiles maps a room's current tile heights into ROOM_HEIGHT_MAP_UPDATE records for a
// specific set of points (e.g. a moved item's old and new footprint), skipping points outside the
// room's row-major tile snapshot.
func HeightMapUpdateTiles(width uint16, tiles []roomlive.TileHeight, points []grid.Point) []outheightmapupdate.Tile {
	seen := make(map[grid.Point]struct{}, len(points))
	records := make([]outheightmapupdate.Tile, 0, len(points))
	for _, point := range points {
		if _, duplicate := seen[point]; duplicate {
			continue
		}
		seen[point] = struct{}{}

		index := int(point.Y)*int(width) + int(point.X)
		if index < 0 || index >= len(tiles) {
			continue
		}
		records = append(records, outheightmapupdate.Tile{X: int(point.X), Y: int(point.Y), Value: heightMapValue(tiles[index])})
	}

	return records
}

// heightMapValue encodes one tile's height and stacking state as a Nitro height-map short.
func heightMapValue(tile roomlive.TileHeight) int16 {
	if !tile.Valid {
		return invalidTileValue
	}

	value := int32(tile.Height) * 256 / int32(grid.HeightScale)
	if tile.StackingBlocked {
		value |= stackingBlockedBit
	}

	return int16(value)
}
