// Package managementfailed contains one Nitro social-group outbound packet.
package managementfailed

import "github.com/niflaot/pixels/networking/codec"

// Header identifies this Nitro packet.
const Header uint16 = 818

// Encode creates the complete packet.
func Encode(groupID int64, reason int32) (codec.Packet, error) {
	return codec.NewPacket(Header, codec.Definition{codec.Int32Field, codec.Int32Field}, codec.Int32(int32(groupID)), codec.Int32(int32(reason)))
}
