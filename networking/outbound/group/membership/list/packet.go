// Package list contains GROUP_LIST.
package list

import (
	grouprecord "github.com/niflaot/pixels/internal/realm/group/record"
	"github.com/niflaot/pixels/networking/codec"
)

// Header identifies GROUP_LIST.
const Header uint16 = 420

// Encode creates Nitro HabboGroupEntryData records.
func Encode(groups []grouprecord.PlayerGroup) (codec.Packet, error) {
	payload, err := codec.AppendPayload(nil, codec.Definition{codec.Int32Field}, codec.Int32(int32(len(groups))))
	if err != nil {
		return codec.Packet{}, err
	}
	for _, item := range groups {
		group := item.Group
		payload, err = codec.AppendPayload(payload, codec.Definition{codec.Int32Field, codec.StringField, codec.StringField, codec.StringField, codec.StringField, codec.BooleanField, codec.Int32Field, codec.BooleanField},
			codec.Int32(int32(group.ID)), codec.String(group.Name), codec.String(group.BadgeCode), codec.String(group.ColorAHex), codec.String(group.ColorBHex), codec.Bool(item.Favorite), codec.Int32(int32(group.OwnerPlayerID)), codec.Bool(group.ForumEnabled))
		if err != nil {
			return codec.Packet{}, err
		}
	}
	return codec.Packet{Header: Header, Payload: payload}, nil
}
