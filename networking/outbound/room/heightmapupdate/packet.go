// Package heightmapupdate contains the ROOM_HEIGHT_MAP_UPDATE outbound packet.
package heightmapupdate

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the ROOM_HEIGHT_MAP_UPDATE packet identifier.
	Header uint16 = 558
)

// Tile stores one changed tile's height and stacking state.
type Tile struct {
	// X stores the tile x coordinate.
	X int

	// Y stores the tile y coordinate.
	Y int

	// Value stores the packed height/stacking wire value, matching ROOM_HEIGHT_MAP's encoding.
	Value int16
}

// Encode creates a ROOM_HEIGHT_MAP_UPDATE packet carrying the tiles a furniture change affected, so
// the client's cached local height map (used for placement and movement prediction) stays in sync
// without waiting for the next full room entry.
func Encode(tiles []Tile) (codec.Packet, error) {
	definition, values := definitionFor(len(tiles))
	values[0] = codec.Byte(uint8(len(tiles)))
	for index, tile := range tiles {
		base := 1 + index*3
		values[base] = codec.Byte(uint8(tile.X))
		values[base+1] = codec.Byte(uint8(tile.Y))
		values[base+2] = codec.Uint16(uint16(tile.Value))
	}

	return codec.NewPacket(Header, definition, values...)
}

// definitionFor builds the tile-count field plus one x/y/value triple per tile.
func definitionFor(tileCount int) (codec.Definition, []codec.Value) {
	definition := make(codec.Definition, 1+tileCount*3)
	definition[0] = codec.Named("count", codec.ByteField)
	for index := range tileCount {
		base := 1 + index*3
		definition[base] = codec.Named("x", codec.ByteField)
		definition[base+1] = codec.Named("y", codec.ByteField)
		definition[base+2] = codec.Named("height", codec.Uint16Field)
	}

	return definition, make([]codec.Value, 1+tileCount*3)
}
