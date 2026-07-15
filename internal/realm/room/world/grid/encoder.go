package grid

import "fmt"

// Encode converts a grid back to the normalized heightmap format.
func Encode(roomGrid Grid) (string, error) {
	if roomGrid.width == 0 || roomGrid.height == 0 {
		return "", ErrEmptyHeightmap
	}

	output := make([]byte, 0, roomGrid.TileCount()+int(roomGrid.height)-1)
	for y := uint16(0); y < roomGrid.height; y++ {
		if y > 0 {
			output = append(output, '\r')
		}
		for x := uint16(0); x < roomGrid.width; x++ {
			value, err := encodeTile(roomGrid, Point{X: x, Y: y})
			if err != nil {
				return "", err
			}
			output = append(output, value)
		}
	}

	return string(output), nil
}

// Encode converts the grid back to the normalized heightmap format.
func (grid Grid) Encode() (string, error) {
	return Encode(grid)
}

// encodeTile converts one grid tile to a heightmap byte.
func encodeTile(roomGrid Grid, point Point) (byte, error) {
	tile, ok := roomGrid.Tile(point)
	if !ok {
		return 0, ErrOutOfBounds
	}
	if !tile.Valid() {
		return 'x', nil
	}

	return encodeHeight(tile.Height())
}

// encodeHeight converts one height value to a heightmap byte.
func encodeHeight(height Height) (byte, error) {
	if height%HeightScale != 0 {
		return 0, fmt.Errorf("%w: %s", ErrInvalidHeight, height.String())
	}
	height /= HeightScale
	if height >= 0 && height <= 9 {
		return byte('0' + height), nil
	}
	if height >= 10 && height <= 35 {
		return byte('a' + height - 10), nil
	}

	return 0, fmt.Errorf("%w: %d", ErrInvalidHeight, height)
}
