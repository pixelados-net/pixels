// Package post contains one Nitro group-forum outbound packet.
package post

import (
	grouprecord "github.com/niflaot/pixels/internal/realm/group/record"
	"github.com/niflaot/pixels/networking/codec"
	forumdata "github.com/niflaot/pixels/networking/outbound/group/forum/data"
	"time"
)

// Header identifies this Nitro packet.
const Header uint16 = 2049

// Encode creates the complete forum update packet.
func Encode(groupID int64, threadID int64, post grouprecord.Post, now time.Time) (codec.Packet, error) {
	payload, err := codec.AppendPayload(nil, codec.Definition{codec.Int32Field, codec.Int32Field}, codec.Int32(int32(groupID)), codec.Int32(int32(threadID)))
	if err != nil {
		return codec.Packet{}, err
	}
	payload, err = forumdata.AppendPost(payload, post, now)
	return codec.Packet{Header: Header, Payload: payload}, err
}
