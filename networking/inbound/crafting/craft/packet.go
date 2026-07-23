// Package craft contains the CRAFT inbound packet.
package craft

import "github.com/niflaot/pixels/networking/codec"

// Header identifies CRAFT.
const Header uint16 = 3591

// Payload stores one named recipe craft request.
type Payload struct {
	// AltarItemID identifies the placed altar instance.
	AltarItemID int64
	// RecipeName identifies the requested visible recipe.
	RecipeName string
}

// Decode reads one named recipe craft request.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, codec.Definition{codec.Int32Field, codec.StringField})
	if err != nil {
		return Payload{}, err
	}
	if values[0].Int32 <= 0 || values[1].String == "" {
		return Payload{}, codec.ErrInvalidField
	}
	return Payload{AltarItemID: int64(values[0].Int32), RecipeName: values[1].String}, nil
}
