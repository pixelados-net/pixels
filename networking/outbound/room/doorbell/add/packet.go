// Package add contains the ROOM_DOORBELL outbound packet.
package add

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the ROOM_DOORBELL packet identifier.
	Header uint16 = 2309
)

// Definition describes the ROOM_DOORBELL payload.
var Definition = codec.Definition{codec.Named("username", codec.StringField)}

// Encode creates a ROOM_DOORBELL packet.
func Encode(username string) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.String(username))
}
