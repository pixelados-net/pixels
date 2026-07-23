// Package denied contains the ROOM_DOORBELL_REJECTED outbound packet.
package denied

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the ROOM_DOORBELL_REJECTED packet identifier.
	Header uint16 = 878
)

// Definition describes the ROOM_DOORBELL_REJECTED payload.
var Definition = codec.Definition{codec.Named("username", codec.StringField)}

// Encode creates a ROOM_DOORBELL_REJECTED packet.
func Encode(username string) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.String(username))
}
