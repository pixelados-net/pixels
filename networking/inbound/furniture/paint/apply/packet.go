// Package apply decodes room surface consumable use.
package apply

import "github.com/niflaot/pixels/networking/codec"

// Header is the ITEM_PAINT identifier.
const Header uint16 = 711

// Payload identifies one room-effect inventory item.
type Payload struct { // ItemID identifies the consumable.
	ItemID int32
}

// Definition describes the wire fields.
var Definition = codec.Definition{codec.Int32Field}

// Decode decodes room surface consumable use.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return Payload{}, err
	}
	return Payload{ItemID: values[0].Int32}, nil
}
