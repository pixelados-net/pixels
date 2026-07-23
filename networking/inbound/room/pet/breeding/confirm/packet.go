// Package confirm decodes PET_CONFIRM_BREEDING requests.
package confirm

import "github.com/niflaot/pixels/networking/codec"

// Header identifies PET_CONFIRM_BREEDING.
const Header uint16 = 3382

// Definition describes PET_CONFIRM_BREEDING fields.
var Definition = codec.Definition{codec.Int32Field, codec.StringField, codec.Int32Field, codec.Int32Field}

// Payload contains one breeding confirmation.
type Payload struct {
	// NestItemID identifies the breeding nest.
	NestItemID int64
	// Name stores the requested offspring name.
	Name string
	// PetOneID identifies the first parent.
	PetOneID int64
	// PetTwoID identifies the second parent.
	PetTwoID int64
}

// Decode decodes PET_CONFIRM_BREEDING.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return Payload{}, err
	}
	return Payload{NestItemID: int64(values[0].Int32), Name: values[1].String, PetOneID: int64(values[2].Int32), PetTwoID: int64(values[3].Int32)}, nil
}
