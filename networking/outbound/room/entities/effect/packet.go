// Package effect encodes UNIT_EFFECT broadcasts.
package effect

import "github.com/niflaot/pixels/networking/codec"

// Header identifies UNIT_EFFECT.
const Header uint16 = 1167

// Definition describes UNIT_EFFECT fields.
var Definition = codec.Definition{codec.Named("roomIndex", codec.Int32Field), codec.Named("effectId", codec.Int32Field), codec.Named("delay", codec.Int32Field)}

// Encode creates a UNIT_EFFECT packet.
func Encode(roomIndex int64, effectID int32, delay int32) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.Int32(int32(roomIndex)), codec.Int32(effectID), codec.Int32(delay))
}
