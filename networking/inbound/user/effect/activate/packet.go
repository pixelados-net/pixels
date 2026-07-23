// Package activate decodes USER_EFFECT_ACTIVATE requests.
package activate

import "github.com/niflaot/pixels/networking/codec"

// Header identifies USER_EFFECT_ACTIVATE.
const Header uint16 = 2959

// Definition describes USER_EFFECT_ACTIVATE fields.
var Definition = codec.Definition{codec.Named("effectId", codec.Int32Field)}

// Decode decodes a USER_EFFECT_ACTIVATE packet.
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
