// Package updated contains the ROOM_THICKNESS outbound packet.
package updated

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header identifies ROOM_THICKNESS.
	Header uint16 = 3547
)

// Definition describes ROOM_THICKNESS fields.
var Definition = codec.Definition{codec.Named("hideWalls", codec.BooleanField), codec.Named("wallThickness", codec.Int32Field), codec.Named("floorThickness", codec.Int32Field)}

// Encode creates a ROOM_THICKNESS packet.
func Encode(hideWalls bool, wallThickness int32, floorThickness int32) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.Bool(hideWalls), codec.Int32(wallThickness), codec.Int32(floorThickness))
}
