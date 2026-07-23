// Package purchased contains one Nitro social-group outbound packet.
package purchased

import "github.com/niflaot/pixels/networking/codec"

// Header identifies this Nitro packet.
const Header uint16 = 2808

// Encode creates the complete packet.
func Encode(roomID int64, groupID int64) (codec.Packet, error) {
	return codec.NewPacket(Header, codec.Definition{codec.Int32Field, codec.Int32Field}, codec.Int32(int32(roomID)), codec.Int32(int32(groupID)))
}
