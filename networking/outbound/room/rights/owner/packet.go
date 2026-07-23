// Package owner contains the ROOM_RIGHTS_OWNER outbound packet.
package owner

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header identifies ROOM_RIGHTS_OWNER.
	Header uint16 = 339
)

// Definition describes the empty ROOM_RIGHTS_OWNER payload.
var Definition = codec.Definition{}

// Encode creates a ROOM_RIGHTS_OWNER packet.
func Encode() (codec.Packet, error) {
	return codec.NewPacket(Header, Definition)
}
