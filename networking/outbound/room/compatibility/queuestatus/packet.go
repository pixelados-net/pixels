// Package queuestatus encodes the renderer-absent ROOM_QUEUE_STATUS compatibility packet.
package queuestatus

import "github.com/niflaot/pixels/networking/codec"

// Header identifies ROOM_QUEUE_STATUS.
const Header uint16 = 2208

// Definition is empty because no renderer parser or reference wire shape exists.
var Definition = codec.Definition{}

// Encode creates the neutral golden-only compatibility packet.
func Encode() (codec.Packet, error) { return codec.NewPacket(Header, Definition) }
