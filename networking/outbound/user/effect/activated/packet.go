// Package activated encodes USER_EFFECT_ACTIVATE acknowledgements.
package activated

import "github.com/niflaot/pixels/networking/codec"

// Header identifies USER_EFFECT_ACTIVATE.
const Header uint16 = 1959

// Definition describes USER_EFFECT_ACTIVATE fields.
var Definition = codec.Definition{codec.Named("effectId", codec.Int32Field), codec.Named("duration", codec.Int32Field), codec.Named("permanent", codec.BooleanField)}

// Encode creates a USER_EFFECT_ACTIVATE packet.
func Encode(effectID int32, duration int32, permanent bool) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.Int32(effectID), codec.Int32(duration), codec.Bool(permanent))
}
