// Package add encodes USER_PET_ADD.
package add

import (
	"github.com/niflaot/pixels/networking/codec"
	petdata "github.com/niflaot/pixels/networking/pet/data"
)

// Header identifies USER_PET_ADD.
const Header uint16 = 2101

// Encode creates USER_PET_ADD.
func Encode(pet petdata.Pet, boughtAsGift bool) (codec.Packet, error) {
	payload, err := petdata.AppendPet(nil, pet)
	if err != nil {
		return codec.Packet{}, err
	}
	payload, err = codec.AppendPayload(payload, codec.Definition{codec.BooleanField}, codec.Bool(boughtAsGift))
	return codec.Packet{Header: Header, Payload: payload}, err
}
