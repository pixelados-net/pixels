// Package move decodes PET_MOVE requests.
package move

import "github.com/niflaot/pixels/networking/codec"

// Header identifies PET_MOVE.
const Header uint16 = 3449

// Definition describes PET_MOVE fields.
var Definition = codec.Definition{codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field}

// Payload contains one directed movement request.
type Payload struct {
	// PetID identifies the room pet.
	PetID int64
	// X stores the destination tile coordinate.
	X int
	// Y stores the destination tile coordinate.
	Y int
	// Direction stores the requested final rotation.
	Direction int
}

// Decode decodes PET_MOVE.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return Payload{}, err
	}
	return Payload{PetID: int64(values[0].Int32), X: int(values[1].Int32), Y: int(values[2].Int32), Direction: int(values[3].Int32)}, nil
}
