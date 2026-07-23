// Package search contains GROUP_MEMBERS.
package search

import (
	grouprecord "github.com/niflaot/pixels/internal/realm/group/record"
	"github.com/niflaot/pixels/networking/codec"
)

// Header identifies GROUP_MEMBERS.
const Header uint16 = 1200

// Encode creates one exact bounded member search page.
func Encode(page grouprecord.MemberPage, canManage bool) (codec.Packet, error) {
	payload, err := codec.AppendPayload(nil, codec.Definition{codec.Int32Field, codec.StringField, codec.Int32Field, codec.StringField, codec.Int32Field, codec.Int32Field},
		codec.Int32(int32(page.Group.ID)), codec.String(page.Group.Name), codec.Int32(int32(page.Group.HomeRoomID)), codec.String(page.Group.BadgeCode), codec.Int32(page.Total), codec.Int32(int32(len(page.Members))))
	if err != nil {
		return codec.Packet{}, err
	}
	for _, member := range page.Members {
		joined := ""
		if member.Role != grouprecord.Requested {
			joined = member.JoinedAt.Format("02/01/2006")
		}
		payload, err = appendMember(payload, member.Role, member.PlayerID, member.Username, member.Figure, joined)
		if err != nil {
			return codec.Packet{}, err
		}
	}
	payload, err = codec.AppendPayload(payload, codec.Definition{codec.BooleanField, codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.StringField}, codec.Bool(canManage), codec.Int32(page.PageSize), codec.Int32(page.Page), codec.Int32(page.Level), codec.String(page.Query))
	return codec.Packet{Header: Header, Payload: payload}, err
}

// appendMember appends one GroupMember wire record.
func appendMember(dst []byte, role grouprecord.Role, playerID int64, username string, figure string, joined string) ([]byte, error) {
	return codec.AppendPayload(dst, codec.Definition{codec.Int32Field, codec.Int32Field, codec.StringField, codec.StringField, codec.StringField}, codec.Int32(int32(role)), codec.Int32(int32(playerID)), codec.String(username), codec.String(figure), codec.String(joined))
}
