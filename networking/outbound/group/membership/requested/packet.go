// Package requested contains GROUP_MEMBERSHIP_REQUESTED.
package requested

import (
	grouprecord "github.com/niflaot/pixels/internal/realm/group/record"
	"github.com/niflaot/pixels/networking/codec"
)

// Header identifies GROUP_MEMBERSHIP_REQUESTED.
const Header uint16 = 1180

// Encode creates one live pending MemberData notification.
func Encode(groupID int64, request grouprecord.Request) (codec.Packet, error) {
	return codec.NewPacket(Header, codec.Definition{codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.StringField, codec.StringField, codec.StringField},
		codec.Int32(int32(groupID)), codec.Int32(2), codec.Int32(int32(request.PlayerID)), codec.String(request.Username), codec.String(request.Figure), codec.String(""))
}
