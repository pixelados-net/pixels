// Package category contains Nitro's inventory unseen-category acknowledgement.
package category

import "github.com/niflaot/pixels/networking/codec"

// Header identifies UNSEEN_RESET_CATEGORY.
const Header uint16 = 3493

// Payload contains one acknowledged inventory category.
type Payload struct {
	// Category identifies the acknowledged inventory category.
	Category int32
}

// Decode unpacks one unseen-category acknowledgement.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, codec.Definition{codec.Int32Field})
	if err != nil {
		return Payload{}, err
	}
	return Payload{Category: values[0].Int32}, nil
}
