// Package unreadcount contains one Nitro social-group outbound packet.
package unreadcount

import "github.com/niflaot/pixels/networking/codec"

// Header identifies this Nitro packet.
const Header uint16 = 2379

// Encode creates the complete packet.
func Encode(count int32) (codec.Packet, error) {
	return codec.NewPacket(Header, codec.Definition{codec.Int32Field}, codec.Int32(int32(count)))
}
