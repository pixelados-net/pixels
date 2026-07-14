// Package enable decodes USER_EFFECT_ENABLE requests.
package enable

import "github.com/niflaot/pixels/networking/codec"

// Header identifies USER_EFFECT_ENABLE.
const Header uint16 = 1752

// Definition describes USER_EFFECT_ENABLE fields.
var Definition = codec.Definition{codec.Named("effectId", codec.Int32Field)}

// Decode decodes a USER_EFFECT_ENABLE packet.
func Decode(packet codec.Packet) (int32, error) {
	if packet.Header != Header {
		return 0, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return 0, err
	}
	return values[0].Int32, nil
}
