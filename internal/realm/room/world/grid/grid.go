package grid

import (
	"math"
	"strconv"
)

// Height stores a room height value in compact fixed-point units.
type Height int16

const (
	// HeightScale stores the number of fixed-point steps in one room height unit.
	HeightScale Height = 4

	// AvatarClearance stores the standing avatar clearance in fixed-point units.
	AvatarClearance Height = 8
)

// HeightFromUnits converts a protocol or persistence height into fixed-point units.
func HeightFromUnits(value float64) Height {
	return Height(math.Round(value * float64(HeightScale)))
}

// HeightFromInt converts a whole room height into fixed-point units.
func HeightFromInt(value int) Height {
	return Height(value) * HeightScale
}

// Units converts a fixed-point height into room units.
func (height Height) Units() float64 {
	return float64(height) / float64(HeightScale)
}

// String returns a protocol-compatible decimal room height.
func (height Height) String() string {
	return strconv.FormatFloat(height.Units(), 'f', -1, 64)
}

// Point stores tile coordinates in a room grid.
type Point struct {
	// X stores the horizontal coordinate.
	X uint16

	// Y stores the vertical coordinate.
	Y uint16
}

// NewPoint creates a point from signed coordinates.
func NewPoint(x int, y int) (Point, bool) {
	if x < 0 || y < 0 || x > int(^uint16(0)) || y > int(^uint16(0)) {
		return Point{}, false
	}

	return Point{X: uint16(x), Y: uint16(y)}, true
}

// MustPoint creates a point and panics when coordinates are invalid.
func MustPoint(x int, y int) Point {
	point, ok := NewPoint(x, y)
	if !ok {
		panic(ErrOutOfBounds)
	}

	return point
}

// PointInFront returns the adjacent point reached by one Habbo rotation.
func PointInFront(point Point, rotation uint8) (Point, bool) {
	dx, dy := directionOffset(rotation)
	return NewPoint(int(point.X)+dx, int(point.Y)+dy)
}

// directionOffset maps a Habbo rotation to its adjacent grid offset.
func directionOffset(rotation uint8) (int, int) {
	switch rotation % 8 {
	case 0:
		return 0, -1
	case 1:
		return 1, -1
	case 2:
		return 1, 0
	case 3:
		return 1, 1
	case 4:
		return 0, 1
	case 5:
		return -1, 1
	case 6:
		return -1, 0
	default:
		return -1, -1
	}
}

// TileFlag stores compact tile metadata.
type TileFlag uint8

const (
	// FlagInvalid marks a tile that does not exist in the room.
	FlagInvalid TileFlag = 1 << iota

	// FlagDoor marks the room door tile.
	FlagDoor
)

// Tile stores a compact immutable view of a room tile.
type Tile struct {
	// point stores the tile coordinate.
	point Point

	// height stores the base tile height.
	height Height

	// flags stores compact tile metadata.
	flags TileFlag
}

// Point returns the tile coordinate.
func (tile Tile) Point() Point {
	return tile.point
}

// Height returns the tile base height.
func (tile Tile) Height() Height {
	return tile.height
}

// Flags returns the tile flags.
func (tile Tile) Flags() TileFlag {
	return tile.flags
}

// Valid reports whether the tile exists.
func (tile Tile) Valid() bool {
	return tile.flags&FlagInvalid == 0
}

// Door reports whether the tile is the room door.
func (tile Tile) Door() bool {
	return tile.flags&FlagDoor != 0
}

// Grid stores an immutable compact room heightmap.
type Grid struct {
	// width stores the number of columns.
	width uint16

	// height stores the number of rows.
	height uint16

	// heights stores base heights indexed by y*width+x.
	heights []Height

	// flags stores tile flags indexed by y*width+x.
	flags []TileFlag

	// door stores the room door coordinate.
	door Point

	// hasDoor reports whether a door was configured.
	hasDoor bool

	// validCount stores the number of existing tiles.
	validCount int
}

// Width returns the grid width.
func (grid Grid) Width() uint16 {
	return grid.width
}

// Height returns the grid height.
func (grid Grid) Height() uint16 {
	return grid.height
}

// TileCount returns the total number of grid tiles.
func (grid Grid) TileCount() int {
	return int(grid.width) * int(grid.height)
}

// ValidCount returns the number of existing tiles.
func (grid Grid) ValidCount() int {
	return grid.validCount
}

// Door returns the room door coordinate.
func (grid Grid) Door() (Point, bool) {
	return grid.door, grid.hasDoor
}

// InBounds reports whether a point belongs to the grid.
func (grid Grid) InBounds(point Point) bool {
	return point.X < grid.width && point.Y < grid.height
}

// Index returns the compact slice index for a point.
func (grid Grid) Index(point Point) (int, bool) {
	if !grid.InBounds(point) {
		return 0, false
	}

	return int(point.Y)*int(grid.width) + int(point.X), true
}

// Tile returns the tile at a point.
func (grid Grid) Tile(point Point) (Tile, bool) {
	index, ok := grid.Index(point)
	if !ok {
		return Tile{}, false
	}

	return Tile{point: point, height: grid.heights[index], flags: grid.flags[index]}, true
}

// HeightAt returns the base height at a point.
func (grid Grid) HeightAt(point Point) (Height, bool) {
	tile, ok := grid.Tile(point)
	if !ok || !tile.Valid() {
		return 0, false
	}

	return tile.Height(), true
}

// FlagsAt returns tile flags at a point.
func (grid Grid) FlagsAt(point Point) (TileFlag, bool) {
	index, ok := grid.Index(point)
	if !ok {
		return 0, false
	}

	return grid.flags[index], true
}

// Valid reports whether a point is inside the grid and exists.
func (grid Grid) Valid(point Point) bool {
	flags, ok := grid.FlagsAt(point)

	return ok && flags&FlagInvalid == 0
}
