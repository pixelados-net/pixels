// Package start decodes PETS_BREED requests.
package start

import "github.com/niflaot/pixels/networking/codec"

// Header identifies PETS_BREED.
const Header uint16 = 1638

// Definition describes PETS_BREED fields.
var Definition = codec.Definition{codec.Int32Field, codec.Int32Field, codec.Int32Field}

// Payload contains one breeding state request.
type Payload struct {
	// State identifies start, cancel, or accept.
	State int32
	// PetOneID identifies the first parent.
	PetOneID int64
	// PetTwoID identifies the second parent.
	PetTwoID int64
}

// Decode decodes PETS_BREED.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return Payload{}, err
	}
	return Payload{State: values[0].Int32, PetOneID: int64(values[1].Int32), PetTwoID: int64(values[2].Int32)}, nil
}
