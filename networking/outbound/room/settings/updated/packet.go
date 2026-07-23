// Package updated contains the ROOM_INFO_UPDATED outbound packet.
package updated

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header identifies ROOM_INFO_UPDATED.
	Header uint16 = 3297
)

// Definition describes ROOM_INFO_UPDATED fields.
var Definition = codec.Definition{codec.Named("roomId", codec.Int32Field)}

// Encode creates a ROOM_INFO_UPDATED packet.
func Encode(roomID int32) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.Int32(roomID))
}
