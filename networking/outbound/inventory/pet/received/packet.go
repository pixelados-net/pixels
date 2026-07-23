// Package received encodes PET_RECEIVED.
package received

import (
	"github.com/niflaot/pixels/networking/codec"
	petdata "github.com/niflaot/pixels/networking/pet/data"
)

// Header identifies PET_RECEIVED.
const Header uint16 = 1111

// Encode creates PET_RECEIVED.
func Encode(boughtAsGift bool, pet petdata.Pet) (codec.Packet, error) {
	payload, err := codec.AppendPayload(nil, codec.Definition{codec.BooleanField}, codec.Bool(boughtAsGift))
	if err != nil {
		return codec.Packet{}, err
	}
	payload, err = petdata.AppendPet(payload, pet)
	return codec.Packet{Header: Header, Payload: payload}, err
}
