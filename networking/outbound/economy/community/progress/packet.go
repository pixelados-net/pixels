// Package progress encodes COMMUNITY_GOAL_PROGRESS responses.
package progress

import "github.com/niflaot/pixels/networking/codec"

// Header identifies COMMUNITY_GOAL_PROGRESS.
const Header uint16 = 2525

// Definition describes the fixed community goal progress prefix and reward count.
var Definition = codec.Definition{
	codec.Named("hasGoalExpired", codec.BooleanField),
	codec.Named("personalContributionScore", codec.Int32Field),
	codec.Named("personalContributionRank", codec.Int32Field),
	codec.Named("communityTotalScore", codec.Int32Field),
	codec.Named("communityHighestAchievedLevel", codec.Int32Field),
	codec.Named("scoreRemainingUntilNextLevel", codec.Int32Field),
	codec.Named("percentCompletionTowardsNextLevel", codec.Int32Field),
	codec.Named("goalCode", codec.StringField),
	codec.Named("timeRemainingInSeconds", codec.Int32Field),
	codec.Named("rewardCount", codec.Int32Field),
}

// RewardLimitDefinition describes one reward level's user limit.
var RewardLimitDefinition = codec.Definition{codec.Named("rewardUserLimit", codec.Int32Field)}

// Encode creates an explicitly expired community goal with no reward levels.
func Encode() (codec.Packet, error) {
	return codec.NewPacket(Header, Definition,
		codec.Bool(true), codec.Int32(0), codec.Int32(0), codec.Int32(0),
		codec.Int32(0), codec.Int32(0), codec.Int32(0), codec.String(""),
		codec.Int32(0), codec.Int32(0))
}
