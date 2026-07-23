// Package list encodes USER_PETS inventory fragments.
package list

import (
	"github.com/niflaot/pixels/networking/codec"
	petdata "github.com/niflaot/pixels/networking/pet/data"
)

// Header identifies USER_PETS.
const Header uint16 = 3522

// Encode creates one ordered pet inventory fragment.
func Encode(totalFragments int32, fragmentNumber int32, pets []petdata.Pet) (codec.Packet, error) {
	payload, err := codec.AppendPayload(nil, codec.Definition{codec.Int32Field, codec.Int32Field, codec.Int32Field},
		codec.Int32(totalFragments), codec.Int32(fragmentNumber), codec.Int32(int32(len(pets))))
	if err != nil {
		return codec.Packet{}, err
	}
	for _, pet := range pets {
		payload, err = petdata.AppendPet(payload, pet)
		if err != nil {
			return codec.Packet{}, err
		}
	}
	return codec.Packet{Header: Header, Payload: payload}, nil
}
