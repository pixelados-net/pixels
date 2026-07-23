// Package get contains the functional GET_CRAFTING_RECIPE inbound packet.
package get

import "github.com/niflaot/pixels/networking/codec"

// Header identifies the string-bearing recipe request despite its crossed legacy name.
const Header uint16 = 633

// Payload stores one recipe name request.
type Payload struct {
	// RecipeName identifies the recipe in the currently opened altar.
	RecipeName string
}

// Decode reads one recipe name request.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, codec.Definition{codec.StringField})
	if err != nil {
		return Payload{}, err
	}
	if values[0].String == "" {
		return Payload{}, codec.ErrInvalidField
	}
	return Payload{RecipeName: values[0].String}, nil
}
