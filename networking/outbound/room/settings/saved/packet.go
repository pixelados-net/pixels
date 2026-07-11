// Package saved contains the ROOM_SETTINGS_SAVE outbound packet.
package saved

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header identifies ROOM_SETTINGS_SAVE.
	Header uint16 = 948
)

// Definition describes ROOM_SETTINGS_SAVE fields.
var Definition = codec.Definition{codec.Named("roomId", codec.Int32Field)}

// Encode creates a ROOM_SETTINGS_SAVE packet.
func Encode(roomID int32) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.Int32(roomID))
}
