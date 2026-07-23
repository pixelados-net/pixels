// Package profile contains the USER_PROFILE outbound packet.
package profile

import (
	grouprecord "github.com/niflaot/pixels/internal/realm/group/record"
	"github.com/niflaot/pixels/networking/codec"
)

// Header identifies USER_PROFILE.
const Header uint16 = 3898

// Encode creates USER_PROFILE with complete social-group memberships.
func Encode(playerID int64, username string, look string, motto string, registration string, groups []grouprecord.PlayerGroup, friendCount int32, isFriend bool, requestSent bool, online bool, secondsSinceLastVisit int32, openWindow bool) (codec.Packet, error) {
	payload, err := codec.AppendPayload(nil, codec.Definition{
		codec.Int32Field, codec.StringField, codec.StringField, codec.StringField, codec.StringField,
		codec.Int32Field, codec.Int32Field, codec.BooleanField, codec.BooleanField, codec.BooleanField, codec.Int32Field,
	}, codec.Int32(int32(playerID)), codec.String(username), codec.String(look), codec.String(motto), codec.String(registration),
		codec.Int32(0), codec.Int32(friendCount), codec.Bool(isFriend), codec.Bool(requestSent), codec.Bool(online), codec.Int32(int32(len(groups))))
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
	payload, err = codec.AppendPayload(payload, codec.Definition{
		codec.Int32Field, codec.BooleanField,
	}, codec.Int32(secondsSinceLastVisit), codec.Bool(openWindow))
	return codec.Packet{Header: Header, Payload: payload}, err
}
