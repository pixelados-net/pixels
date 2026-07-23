// Package new contains the retired MESSENGER_MINIMAIL_NEW outbound packet.
package new

import "github.com/niflaot/pixels/networking/codec"

// Header identifies MESSENGER_MINIMAIL_NEW.
const Header uint16 = 1911

// Definition describes the empty MESSENGER_MINIMAIL_NEW payload.
var Definition = codec.Definition{}

// Encode creates the retired MESSENGER_MINIMAIL_NEW packet.
//
// Deprecated: MiniMail is retired and intentionally has no runtime integration.
func Encode() (codec.Packet, error) {
	return codec.NewPacket(Header, Definition)
}
