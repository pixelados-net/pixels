// Package selected encodes AVATAR_EFFECT_SELECTED acknowledgements.
package selected

import "github.com/niflaot/pixels/networking/codec"

// Header identifies AVATAR_EFFECT_SELECTED.
const Header uint16 = 3473

// Definition describes AVATAR_EFFECT_SELECTED fields.
var Definition = codec.Definition{codec.Named("effectId", codec.Int32Field)}

// Encode creates an AVATAR_EFFECT_SELECTED packet.
func Encode(effectID int32) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.Int32(effectID))
}
