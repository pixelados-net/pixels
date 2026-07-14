// Package remove encodes USER_EFFECT_LIST_REMOVE updates.
package remove

import "github.com/niflaot/pixels/networking/codec"

// Header identifies USER_EFFECT_LIST_REMOVE.
const Header uint16 = 2228

// Definition describes USER_EFFECT_LIST_REMOVE fields.
var Definition = codec.Definition{codec.Named("effectId", codec.Int32Field)}

// Encode creates a USER_EFFECT_LIST_REMOVE packet.
func Encode(effectID int32) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.Int32(effectID))
}
