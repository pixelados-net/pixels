// Package concurrentprogress encodes CONCURRENT_USERS_GOAL_PROGRESS responses.
package concurrentprogress

import "github.com/niflaot/pixels/networking/codec"

// Header identifies CONCURRENT_USERS_GOAL_PROGRESS.
const Header uint16 = 2737

// Definition describes concurrent users goal state and counts.
var Definition = codec.Definition{
	codec.Named("state", codec.Int32Field),
	codec.Named("userCount", codec.Int32Field),
	codec.Named("userCountGoal", codec.Int32Field),
}

// Encode creates one concurrent users goal snapshot.
func Encode(state int32, userCount int32, userCountGoal int32) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.Int32(state), codec.Int32(userCount), codec.Int32(userCountGoal))
}
