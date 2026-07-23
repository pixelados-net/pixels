// Package friends encodes WEEKLY_GAME2_FRIENDS_LEADERBOARD responses.
package friends

import "github.com/niflaot/pixels/networking/codec"

// Header identifies WEEKLY_GAME2_FRIENDS_LEADERBOARD.
const Header uint16 = 2270

// Encode creates the complete leaderboard metadata shape consumed by Nitro.
func Encode(year int32, week int32, maxOffset int32, currentOffset int32, minutesUntilReset int32) (codec.Packet, error) {
	return codec.NewPacket(Header, codec.Definition{codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field}, codec.Int32(year), codec.Int32(week), codec.Int32(maxOffset), codec.Int32(currentOffset), codec.Int32(minutesUntilReset))
}
