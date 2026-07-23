// Package contextmenu encodes BOT_FORCE_OPEN_CONTEXT_MENU.
package contextmenu

import "github.com/niflaot/pixels/networking/codec"

// Header identifies BOT_FORCE_OPEN_CONTEXT_MENU.
const Header uint16 = 296

// Encode creates BOT_FORCE_OPEN_CONTEXT_MENU.
func Encode(botID int64) (codec.Packet, error) {
	return codec.NewPacket(Header, codec.Definition{codec.Int32Field}, codec.Int32(int32(-botID)))
}
