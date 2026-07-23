// Package refresh contains one Nitro social-group outbound packet.
package refresh

import "github.com/niflaot/pixels/networking/codec"

// Header identifies this Nitro packet.
const Header uint16 = 2445

// Encode creates the complete packet.
func Encode(groupID int64) (codec.Packet, error) {
	return codec.NewPacket(Header, codec.Definition{codec.Int32Field}, codec.Int32(int32(groupID)))
}
