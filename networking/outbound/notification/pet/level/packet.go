// Package level encodes PET_LEVEL_NOTIFICATION.
package level

import (
	"github.com/niflaot/pixels/networking/codec"
	petdata "github.com/niflaot/pixels/networking/pet/data"
)

// Header identifies PET_LEVEL_NOTIFICATION.
const Header uint16 = 859

// Encode creates PET_LEVEL_NOTIFICATION.
func Encode(petID int64, name string, level int32, figure petdata.Figure) (codec.Packet, error) {
	payload, err := codec.AppendPayload(nil, codec.Definition{codec.Int32Field, codec.StringField, codec.Int32Field}, codec.Int32(int32(petID)), codec.String(name), codec.Int32(level))
	if err != nil {
		return codec.Packet{}, err
	}
	payload, err = petdata.AppendFigure(payload, figure)
	return codec.Packet{Header: Header, Payload: payload}, err
}
