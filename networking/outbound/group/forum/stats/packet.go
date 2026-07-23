// Package stats contains GROUP_FORUM_DATA.
package stats

import (
	grouprecord "github.com/niflaot/pixels/internal/realm/group/record"
	"github.com/niflaot/pixels/networking/codec"
	forumdata "github.com/niflaot/pixels/networking/outbound/group/forum/data"
	"time"
)

// Header identifies GROUP_FORUM_DATA.
const Header uint16 = 3011

// Encode creates ExtendedForumData with already-localized errors.
func Encode(summary grouprecord.ForumSummary, errors [5]string, canChangeSettings bool, staff bool, now time.Time) (codec.Packet, error) {
	payload, err := forumdata.AppendSummary(nil, summary, now)
	if err != nil {
		return codec.Packet{}, err
	}
	group := summary.Group
	payload, err = codec.AppendPayload(payload, codec.Definition{
		codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field,
		codec.StringField, codec.StringField, codec.StringField, codec.StringField, codec.StringField,
		codec.BooleanField, codec.BooleanField,
	}, codec.Int32(int32(group.ReadPolicy)), codec.Int32(int32(group.PostMessagePolicy)), codec.Int32(int32(group.PostThreadPolicy)), codec.Int32(int32(group.ModeratePolicy)),
		codec.String(errors[0]), codec.String(errors[1]), codec.String(errors[2]), codec.String(errors[3]), codec.String(errors[4]), codec.Bool(canChangeSettings), codec.Bool(staff))
	return codec.Packet{Header: Header, Payload: payload}, err
}
