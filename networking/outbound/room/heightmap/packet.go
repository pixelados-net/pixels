// Package heightmap contains the ROOM_HEIGHT_MAP outbound packet.
package heightmap

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the ROOM_HEIGHT_MAP packet identifier.
	Header uint16 = 2753
)

// Encode creates a ROOM_HEIGHT_MAP packet from row-major tile values.
func Encode(width int, values []int16) (codec.Packet, error) {
	definition, packetValues := definitionFor(len(values))
	packetValues[0] = codec.Int32(int32(width))
	packetValues[1] = codec.Int32(int32(len(values)))
	for index, value := range values {
		packetValues[2+index] = codec.Uint16(uint16(value))
	}

	return codec.NewPacket(Header, definition, packetValues...)
}

// definitionFor builds the width/tile-count header fields plus one field per tile.
func definitionFor(tileCount int) (codec.Definition, []codec.Value) {
	definition := make(codec.Definition, 2+tileCount)
	definition[0] = codec.Named("width", codec.Int32Field)
	definition[1] = codec.Named("totalTiles", codec.Int32Field)
	for index := range tileCount {
		definition[2+index] = codec.Named("height", codec.Uint16Field)
	}

	return definition, make([]codec.Value, 2+tileCount)
}
