// Package remove encodes USER_BOT_REMOVE.
package remove

import "github.com/niflaot/pixels/networking/codec"

// Header identifies USER_BOT_REMOVE.
const Header uint16 = 233

// Encode creates USER_BOT_REMOVE.
func Encode(botID int64) (codec.Packet, error) {
	return codec.NewPacket(Header, codec.Definition{codec.Int32Field}, codec.Int32(int32(botID)))
}
