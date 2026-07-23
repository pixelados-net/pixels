// Package idle encodes UNIT_IDLE broadcasts.
package idle

import "github.com/niflaot/pixels/networking/codec"

// Header identifies UNIT_IDLE.
const Header uint16 = 1797

// Definition describes UNIT_IDLE fields.
var Definition = codec.Definition{codec.Named("roomIndex", codec.Int32Field), codec.Named("idle", codec.BooleanField)}

// Encode creates a UNIT_IDLE packet.
func Encode(roomIndex int64, value bool) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.Int32(int32(roomIndex)), codec.Bool(value))
}
