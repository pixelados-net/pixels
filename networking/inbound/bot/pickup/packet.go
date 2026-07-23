// Package pickup decodes BOT_PICKUP requests.
package pickup

import "github.com/niflaot/pixels/networking/codec"

// Header identifies BOT_PICKUP.
const Header uint16 = 3323

// Definition describes BOT_PICKUP fields.
var Definition = codec.Definition{codec.Named("botId", codec.Int32Field)}

// Decode decodes BOT_PICKUP.
func Decode(packet codec.Packet) (int64, error) {
	if packet.Header != Header {
		return 0, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return 0, err
	}
	value := int64(values[0].Int32)
	if value < 0 {
		value = -value
	}
	return value, nil
}
