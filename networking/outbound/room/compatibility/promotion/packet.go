// Package promotion encodes renderer-absent ROOM_PROMOTION compatibility packets.
package promotion

import "github.com/niflaot/pixels/networking/codec"

// Header identifies ROOM_PROMOTION.
const Header uint16 = 2274

// Definition is empty because the installed renderer registers no parser.
var Definition = codec.Definition{}

// Encode creates the neutral golden-only ROOM_PROMOTION packet.
func Encode() (codec.Packet, error) { return codec.NewPacket(Header, Definition) }
