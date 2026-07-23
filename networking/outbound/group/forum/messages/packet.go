// Package messages contains GROUP_FORUM_THREAD_MESSAGES.
package messages

import (
	grouprecord "github.com/niflaot/pixels/internal/realm/group/record"
	"github.com/niflaot/pixels/networking/codec"
	forumdata "github.com/niflaot/pixels/networking/outbound/group/forum/data"
	"time"
)

// Header identifies GROUP_FORUM_THREAD_MESSAGES.
const Header uint16 = 509

// Encode creates one bounded MessageData page.
func Encode(groupID int64, threadID int64, start int32, posts []grouprecord.Post, now time.Time) (codec.Packet, error) {
	payload, err := codec.AppendPayload(nil, codec.Definition{codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field}, codec.Int32(int32(groupID)), codec.Int32(int32(threadID)), codec.Int32(start), codec.Int32(int32(len(posts))))
	if err != nil {
		return codec.Packet{}, err
	}
	for _, post := range posts {
		payload, err = forumdata.AppendPost(payload, post, now)
		if err != nil {
			return codec.Packet{}, err
		}
	}
	return codec.Packet{Header: Header, Payload: payload}, nil
}
