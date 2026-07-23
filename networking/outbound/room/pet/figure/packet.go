// Package figure encodes PET_FIGURE_UPDATE.
package figure

import (
	"github.com/niflaot/pixels/networking/codec"
	petdata "github.com/niflaot/pixels/networking/pet/data"
)

// Header identifies PET_FIGURE_UPDATE.
const Header uint16 = 1924

// Encode creates PET_FIGURE_UPDATE.
func Encode(roomIndex int64, petID int64, value petdata.Figure, hasSaddle bool, riding bool) (codec.Packet, error) {
	payload, err := codec.AppendPayload(nil, codec.Definition{codec.Int32Field, codec.Int32Field}, codec.Int32(int32(roomIndex)), codec.Int32(int32(petID)))
	if err != nil {
		return codec.Packet{}, err
	}
	payload, err = petdata.AppendFigure(payload, value)
	if err != nil {
		return codec.Packet{}, err
	}
	payload, err = codec.AppendPayload(payload, codec.Definition{codec.BooleanField, codec.BooleanField}, codec.Bool(hasSaddle), codec.Bool(riding))
	return codec.Packet{Header: Header, Payload: payload}, err
}
