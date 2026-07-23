// Package competitive encodes WEEKLY_COMPETITIVE_LEADERBOARD responses.
package competitive

import "github.com/niflaot/pixels/networking/codec"

// Header identifies WEEKLY_COMPETITIVE_LEADERBOARD.
const Header uint16 = 3512

// Encode creates the complete leaderboard metadata shape consumed by Nitro.
func Encode(year int32, week int32, maxOffset int32, currentOffset int32, minutesUntilReset int32) (codec.Packet, error) {
	return codec.NewPacket(Header, codec.Definition{codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field}, codec.Int32(year), codec.Int32(week), codec.Int32(maxOffset), codec.Int32(currentOffset), codec.Int32(minutesUntilReset))
}
