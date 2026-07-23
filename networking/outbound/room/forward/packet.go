// Package forward contains the ROOM_FORWARD outbound packet.
package forward

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the ROOM_FORWARD packet identifier.
	Header uint16 = 160
)

// Definition describes the ROOM_FORWARD payload fields.
var Definition = codec.Definition{codec.Named("roomId", codec.Int32Field)}

// Encode creates a ROOM_FORWARD packet.
func Encode(roomID int32) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.Int32(roomID))
}
