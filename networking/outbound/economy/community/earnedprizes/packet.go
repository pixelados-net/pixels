// Package earnedprizes encodes COMMUNITY_GOAL_EARNED_PRIZES responses.
package earnedprizes

import "github.com/niflaot/pixels/networking/codec"

// Header identifies COMMUNITY_GOAL_EARNED_PRIZES.
const Header uint16 = 3319

// Definition describes the earned community prize count.
var Definition = codec.Definition{codec.Named("prizeCount", codec.Int32Field)}

// PrizeDefinition describes one earned community goal prize.
var PrizeDefinition = codec.Definition{
	codec.Named("communityGoalId", codec.Int32Field),
	codec.Named("communityGoalCode", codec.StringField),
	codec.Named("userRank", codec.Int32Field),
	codec.Named("rewardCode", codec.StringField),
	codec.Named("badge", codec.BooleanField),
	codec.Named("localizedName", codec.StringField),
}

// Encode creates an explicit empty earned-prize snapshot.
func Encode() (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.Int32(0))
}
