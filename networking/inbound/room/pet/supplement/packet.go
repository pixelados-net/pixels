// Package supplement decodes PET_SUPPLEMENT requests.
package supplement

import "github.com/niflaot/pixels/networking/codec"

// Header identifies PET_SUPPLEMENT.
const Header uint16 = 749

// Definition describes PET_SUPPLEMENT fields.
var Definition = codec.Definition{codec.Int32Field, codec.Int32Field}

// Payload contains one supplement request.
type Payload struct {
	// PetID identifies the pet.
	PetID int64
	// Type identifies the supplement kind.
	Type int32
}

// Decode decodes PET_SUPPLEMENT.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return Payload{}, err
	}
	return Payload{PetID: int64(values[0].Int32), Type: values[1].Int32}, nil
}
