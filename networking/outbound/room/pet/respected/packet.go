// Package respected encodes PET_RESPECTED.
package respected

import (
	"github.com/niflaot/pixels/networking/codec"
	petdata "github.com/niflaot/pixels/networking/pet/data"
)

// Header identifies PET_RESPECTED.
const Header uint16 = 2788

// Encode creates PET_RESPECTED.
func Encode(respect int32, ownerID int64, pet petdata.Pet) (codec.Packet, error) {
	payload, err := codec.AppendPayload(nil, codec.Definition{codec.Int32Field, codec.Int32Field}, codec.Int32(respect), codec.Int32(int32(ownerID)))
	if err != nil {
		return codec.Packet{}, err
	}
	payload, err = petdata.AppendPet(payload, pet)
	return codec.Packet{Header: Header, Payload: payload}, err
}
