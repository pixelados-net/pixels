// Package entered contains the ROOM_ENTER outbound packet.
package entered

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the ROOM_ENTER packet identifier.
	Header uint16 = 758
)

// Definition describes the ROOM_ENTER payload fields.
var Definition = codec.Definition{}

// Encode creates a ROOM_ENTER packet.
func Encode() (codec.Packet, error) {
	return codec.NewPacket(Header, Definition)
}
