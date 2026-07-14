// Package add encodes USER_EFFECT_LIST_ADD updates.
package add

import "github.com/niflaot/pixels/networking/codec"

// Header identifies USER_EFFECT_LIST_ADD.
const Header uint16 = 2867

// Definition describes USER_EFFECT_LIST_ADD fields.
var Definition = codec.Definition{codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.BooleanField}

// Encode creates a USER_EFFECT_LIST_ADD packet.
func Encode(effectID int32, subtype int32, duration int32, inactive int32, secondsLeft int32, permanent bool) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.Int32(effectID), codec.Int32(subtype), codec.Int32(duration), codec.Int32(inactive), codec.Int32(secondsLeft), codec.Bool(permanent))
}
