// Package weekly encodes WEEKLY_GAME_REWARD responses.
package weekly

import "github.com/niflaot/pixels/networking/codec"

// Header identifies WEEKLY_GAME_REWARD.
const Header uint16 = 2641

// EncodeEmpty creates a valid response without catalog products.
func EncodeEmpty(gameTypeID int32, minutesUntilNextWeek int32, rewarding bool) (codec.Packet, error) {
	return codec.NewPacket(Header, codec.Definition{codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.BooleanField}, codec.Int32(gameTypeID), codec.Int32(0), codec.Int32(minutesUntilNextWeek), codec.Bool(rewarding))
}
