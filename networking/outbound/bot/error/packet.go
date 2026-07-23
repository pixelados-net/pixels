// Package error encodes BOT_ERROR.
package error

import "github.com/niflaot/pixels/networking/codec"

// Header identifies BOT_ERROR.
const Header uint16 = 639

// Encode creates BOT_ERROR.
func Encode(code int32) (codec.Packet, error) {
	return codec.NewPacket(Header, codec.Definition{codec.Int32Field}, codec.Int32(code))
}
