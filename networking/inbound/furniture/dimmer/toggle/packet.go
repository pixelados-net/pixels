// Package toggle decodes a mood-light toggle request.
package toggle

import "github.com/niflaot/pixels/networking/codec"

// Header is the ITEM_DIMMER_TOGGLE identifier.
const Header uint16 = 2296

// Payload contains no wire fields.
type Payload struct{}

// Definition describes the empty payload.
var Definition = codec.Definition{}

// Decode decodes a mood-light toggle request.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	_, err := codec.DecodePacketExact(packet, Definition)
	return Payload{}, err
}
