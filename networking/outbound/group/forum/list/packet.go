// Package list contains GROUP_FORUM_LIST.
package list

import (
	grouprecord "github.com/niflaot/pixels/internal/realm/group/record"
	"github.com/niflaot/pixels/networking/codec"
	forumdata "github.com/niflaot/pixels/networking/outbound/group/forum/data"
	"time"
)

// Header identifies GROUP_FORUM_LIST.
const Header uint16 = 3001

// Encode creates one bounded ForumData list page.
func Encode(mode int32, total int32, start int32, summaries []grouprecord.ForumSummary, now time.Time) (codec.Packet, error) {
	payload, err := codec.AppendPayload(nil, codec.Definition{codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field}, codec.Int32(mode), codec.Int32(total), codec.Int32(start), codec.Int32(int32(len(summaries))))
	if err != nil {
		return codec.Packet{}, err
	}
	for _, summary := range summaries {
		payload, err = forumdata.AppendSummary(payload, summary, now)
		if err != nil {
			return codec.Packet{}, err
		}
	}
	return codec.Packet{Header: Header, Payload: payload}, nil
}
