// Package list contains GROUP_BADGES.
package list

import (
	grouprecord "github.com/niflaot/pixels/internal/realm/group/record"
	"github.com/niflaot/pixels/networking/codec"
)

// Header identifies GROUP_BADGES.
const Header uint16 = 2402

// Encode creates relevant group badge identifier pairs.
func Encode(groups []grouprecord.PlayerGroup) (codec.Packet, error) {
	payload, err := codec.AppendPayload(nil, codec.Definition{codec.Int32Field}, codec.Int32(int32(len(groups))))
	if err != nil {
		return codec.Packet{}, err
	}
	for _, item := range groups {
		payload, err = codec.AppendPayload(payload, codec.Definition{codec.Int32Field, codec.StringField}, codec.Int32(int32(item.Group.ID)), codec.String(item.Group.BadgeCode))
		if err != nil {
			return codec.Packet{}, err
		}
	}
	return codec.Packet{Header: Header, Payload: payload}, nil
}
