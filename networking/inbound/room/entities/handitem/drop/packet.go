// Package drop decodes dropping a carried hand item.
package drop

import "github.com/niflaot/pixels/networking/codec"

// Header is the UNIT_DROP_HAND_ITEM identifier.
const Header uint16 = 2814

// Payload contains no wire fields.
type Payload struct{}

// Definition describes the empty payload.
var Definition = codec.Definition{}

// Decode decodes dropping a carried hand item.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	_, err := codec.DecodePacketExact(packet, Definition)
	return Payload{}, err
}
