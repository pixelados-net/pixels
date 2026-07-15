// Package apply decodes a background-toner update.
package apply

import "github.com/niflaot/pixels/networking/codec"

// Header is the ROOM_TONER_APPLY identifier.
const Header uint16 = 2880

// Payload contains toner item and HSL values.
type Payload struct {
	// ItemID identifies the toner furniture.
	ItemID int32
	// Hue stores eight-bit hue.
	Hue int32
	// Saturation stores eight-bit saturation.
	Saturation int32
	// Lightness stores eight-bit lightness.
	Lightness int32
}

// Definition describes the wire fields.
var Definition = codec.Definition{codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field}

// Decode decodes a toner update.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return Payload{}, err
	}
	return Payload{ItemID: values[0].Int32, Hue: values[1].Int32, Saturation: values[2].Int32, Lightness: values[3].Int32}, nil
}
