// Package pickup decodes PET_PICKUP requests.
package pickup

import "github.com/niflaot/pixels/networking/codec"

// Header identifies PET_PICKUP.
const Header uint16 = 1581

// Decode decodes the requested pet identifier.
func Decode(packet codec.Packet) (int64, error) {
	if packet.Header != Header {
		return 0, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, codec.Definition{codec.Int32Field})
	if err != nil {
		return 0, err
	}
	return int64(values[0].Int32), nil
}
