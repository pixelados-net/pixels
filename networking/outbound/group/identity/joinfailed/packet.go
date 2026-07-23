// Package joinfailed contains one Nitro social-group outbound packet.
package joinfailed

import "github.com/niflaot/pixels/networking/codec"

// Header identifies this Nitro packet.
const Header uint16 = 762

// Encode creates the complete packet.
func Encode(reason int32) (codec.Packet, error) {
	return codec.NewPacket(Header, codec.Definition{codec.Int32Field}, codec.Int32(int32(reason)))
}
