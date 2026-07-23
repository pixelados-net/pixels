// Package entrytile contains the ROOM_MODEL_DOOR outbound packet.
package entrytile

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the ROOM_MODEL_DOOR packet identifier.
	Header uint16 = 1664
)

// Definition describes the ROOM_MODEL_DOOR payload fields.
var Definition = codec.Definition{
	codec.Named("x", codec.Int32Field),
	codec.Named("y", codec.Int32Field),
	codec.Named("direction", codec.Int32Field),
}

// Encode creates a ROOM_MODEL_DOOR packet.
func Encode(x int32, y int32, direction int32) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.Int32(x), codec.Int32(y), codec.Int32(direction))
}
