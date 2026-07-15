// Package get decodes a post-it data request.
package get

import "github.com/niflaot/pixels/networking/codec"

// Header is the GET_ITEM_DATA identifier.
const Header uint16 = 3964

// Payload identifies one wall item.
type Payload struct { // ItemID identifies the wall item.
	ItemID int32
}

// Definition describes the wire fields.
var Definition = codec.Definition{codec.Int32Field}

// Decode decodes a post-it data request.
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
