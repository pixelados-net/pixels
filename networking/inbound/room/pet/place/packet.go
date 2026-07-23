// Package place decodes PET_PLACE requests.
package place

import "github.com/niflaot/pixels/networking/codec"

// Header identifies PET_PLACE.
const Header uint16 = 2647

// Definition describes PET_PLACE fields.
var Definition = codec.Definition{codec.Named("petId", codec.Int32Field), codec.Named("x", codec.Int32Field), codec.Named("y", codec.Int32Field)}

// Payload contains one placement request.
type Payload struct {
	// PetID identifies the inventory pet.
	PetID int64
	// X stores the requested tile coordinate.
	X int
	// Y stores the requested tile coordinate.
	Y int
}

// Decode decodes PET_PLACE.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return Payload{}, err
	}
	return Payload{PetID: int64(values[0].Int32), X: int(values[1].Int32), Y: int(values[2].Int32)}, nil
}
