// Package use decodes USE_PET_PRODUCT requests.
package use

import "github.com/niflaot/pixels/networking/codec"

// Header identifies USE_PET_PRODUCT.
const Header uint16 = 1328

// Definition describes USE_PET_PRODUCT fields.
var Definition = codec.Definition{codec.Int32Field, codec.Int32Field}

// Payload contains one product use request.
type Payload struct {
	// ItemID identifies the furniture product.
	ItemID int64
	// PetID identifies the target pet.
	PetID int64
}

// Decode decodes USE_PET_PRODUCT.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return Payload{}, err
	}
	return Payload{ItemID: int64(values[0].Int32), PetID: int64(values[1].Int32)}, nil
}
