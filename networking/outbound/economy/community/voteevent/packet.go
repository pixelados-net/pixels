// Package voteevent encodes COMMUNITY_GOAL_VOTE_EVENT responses.
package voteevent

import "github.com/niflaot/pixels/networking/codec"

// Header identifies COMMUNITY_GOAL_VOTE_EVENT.
const Header uint16 = 1435

// Definition describes the community goal vote acknowledgement.
var Definition = codec.Definition{codec.Named("acknowledged", codec.BooleanField)}

// Encode creates one community goal vote acknowledgement.
func Encode(acknowledged bool) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.Bool(acknowledged))
}
