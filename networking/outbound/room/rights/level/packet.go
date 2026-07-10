// Package level contains the ROOM_RIGHTS outbound packet.
package level

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header identifies ROOM_RIGHTS.
	Header uint16 = 780
	// None identifies no room control.
	None int32 = 0
	// Rights identifies room rights control.
	Rights int32 = 1
	// Owner identifies room owner control.
	Owner int32 = 2
)

// Definition describes ROOM_RIGHTS fields.
var Definition = codec.Definition{codec.Named("level", codec.Int32Field)}

// Encode creates a ROOM_RIGHTS packet.
func Encode(value int32) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.Int32(value))
}
