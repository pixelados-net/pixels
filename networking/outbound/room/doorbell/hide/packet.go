// Package hide contains the ROOM_DOORBELL_ACCEPTED outbound packet.
package hide

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the ROOM_DOORBELL_ACCEPTED packet identifier.
	Header uint16 = 3783
)

// Definition describes the ROOM_DOORBELL_ACCEPTED payload.
var Definition = codec.Definition{codec.Named("username", codec.StringField)}

// Encode creates a ROOM_DOORBELL_ACCEPTED packet.
func Encode(username string) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.String(username))
}
