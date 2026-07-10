// Package clear contains the ROOM_RIGHTS_CLEAR outbound packet.
package clear

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header identifies ROOM_RIGHTS_CLEAR.
	Header uint16 = 2392
)

// Definition describes the empty ROOM_RIGHTS_CLEAR payload.
var Definition = codec.Definition{}

// Encode creates a ROOM_RIGHTS_CLEAR packet.
func Encode() (codec.Packet, error) {
	return codec.NewPacket(Header, Definition)
}
