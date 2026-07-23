// Package options contains GROUP_CREATE_OPTIONS.
package options

import (
	grouprecord "github.com/niflaot/pixels/internal/realm/group/record"
	"github.com/niflaot/pixels/networking/codec"
)

// Header identifies GROUP_CREATE_OPTIONS.
const Header uint16 = 2159

// Encode creates creator cost and eligible room data.
func Encode(cost int64, rooms []grouprecord.EligibleRoom) (codec.Packet, error) {
	payload, err := codec.AppendPayload(nil, codec.Definition{codec.Int32Field, codec.Int32Field}, codec.Int32(int32(cost)), codec.Int32(int32(len(rooms))))
	if err != nil {
		return codec.Packet{}, err
	}
	for _, room := range rooms {
		payload, err = codec.AppendPayload(payload, codec.Definition{codec.Int32Field, codec.StringField, codec.BooleanField}, codec.Int32(int32(room.ID)), codec.String(room.Name), codec.Bool(false))
		if err != nil {
			return codec.Packet{}, err
		}
	}
	return codec.Packet{Header: Header, Payload: payload}, nil
}
