// Package threads contains GROUP_FORUM_THREADS.
package threads

import (
	grouprecord "github.com/niflaot/pixels/internal/realm/group/record"
	"github.com/niflaot/pixels/networking/codec"
	forumdata "github.com/niflaot/pixels/networking/outbound/group/forum/data"
	"time"
)

// Header identifies GROUP_FORUM_THREADS.
const Header uint16 = 1073

// Encode creates one bounded GuildForumThread page.
func Encode(groupID int64, start int32, threads []grouprecord.Thread, now time.Time) (codec.Packet, error) {
	payload, err := codec.AppendPayload(nil, codec.Definition{codec.Int32Field, codec.Int32Field, codec.Int32Field}, codec.Int32(int32(groupID)), codec.Int32(start), codec.Int32(int32(len(threads))))
	if err != nil {
		return codec.Packet{}, err
	}
	for _, thread := range threads {
		payload, err = forumdata.AppendThread(payload, thread, now)
		if err != nil {
			return codec.Packet{}, err
		}
	}
	return codec.Packet{Header: Header, Payload: payload}, nil
}
