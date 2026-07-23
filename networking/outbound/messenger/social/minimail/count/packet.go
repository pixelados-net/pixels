// Package count contains the retired MESSENGER_MINIMAIL_COUNT outbound packet.
package count

import "github.com/niflaot/pixels/networking/codec"

// Header identifies MESSENGER_MINIMAIL_COUNT.
const Header uint16 = 2803

// Definition describes the historical MiniMail unread count.
var Definition = codec.Definition{codec.Named("unreadCount", codec.Int32Field)}

// Encode creates the retired MESSENGER_MINIMAIL_COUNT packet.
//
// Deprecated: MiniMail is retired and intentionally has no runtime integration.
func Encode(unreadCount int32) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.Int32(unreadCount))
}
