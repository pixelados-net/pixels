// Package open contains the NAVIGATOR_OPEN_ROOM_CREATOR outbound packet.
package open

import "github.com/niflaot/pixels/networking/codec"

// Header identifies NAVIGATOR_OPEN_ROOM_CREATOR.
const Header uint16 = 2064

// Encode creates a header-only NAVIGATOR_OPEN_ROOM_CREATOR packet.
func Encode() (codec.Packet, error) { return codec.Packet{Header: Header}, nil }
