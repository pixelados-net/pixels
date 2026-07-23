// Package settings contains GROUP_SETTINGS.
package settings

import (
	grouprecord "github.com/niflaot/pixels/internal/realm/group/record"
	"github.com/niflaot/pixels/networking/codec"
)

// Header identifies GROUP_SETTINGS.
const Header uint16 = 3965

// Encode creates exact manager data including five padded badge layers.
func Encode(group grouprecord.Group, parts []grouprecord.BadgePart) (codec.Packet, error) {
	decorate := int32(1)
	if group.CanMembersDecorate {
		decorate = 0
	}
	payload, err := codec.AppendPayload(nil, codec.Definition{
		codec.Int32Field, codec.Int32Field, codec.StringField, codec.BooleanField, codec.BooleanField,
		codec.Int32Field, codec.StringField, codec.StringField, codec.Int32Field, codec.Int32Field, codec.Int32Field,
		codec.Int32Field, codec.Int32Field, codec.BooleanField, codec.StringField, codec.Int32Field,
	}, codec.Int32(1), codec.Int32(int32(group.HomeRoomID)), codec.String(group.HomeRoomName), codec.Bool(false), codec.Bool(true),
		codec.Int32(int32(group.ID)), codec.String(group.Name), codec.String(group.Description), codec.Int32(int32(group.HomeRoomID)), codec.Int32(group.ColorA), codec.Int32(group.ColorB),
		codec.Int32(int32(group.State)), codec.Int32(decorate), codec.Bool(false), codec.String(""), codec.Int32(5))
	if err != nil {
		return codec.Packet{}, err
	}
	for index := 0; index < 5; index++ {
		part := grouprecord.BadgePart{}
		if index < len(parts) {
			part = parts[index]
		}
		payload, err = codec.AppendPayload(payload, codec.Definition{codec.Int32Field, codec.Int32Field, codec.Int32Field}, codec.Int32(part.ElementID), codec.Int32(part.ColorID), codec.Int32(part.Position))
		if err != nil {
			return codec.Packet{}, err
		}
	}
	payload, err = codec.AppendPayload(payload, codec.Definition{codec.StringField, codec.Int32Field}, codec.String(group.BadgeCode), codec.Int32(group.MemberCount))
	return codec.Packet{Header: Header, Payload: payload}, err
}
