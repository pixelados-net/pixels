// Package member contains one Nitro social-group outbound packet.
package member

import "github.com/niflaot/pixels/networking/codec"

// Header identifies this Nitro packet.
const Header uint16 = 265

// Encode creates the complete packet.
func Encode(groupID int64, playerID int64, active bool) (codec.Packet, error) {
	return codec.NewPacket(Header, codec.Definition{codec.Int32Field, codec.Int32Field, codec.BooleanField}, codec.Int32(int32(groupID)), codec.Int32(int32(playerID)), codec.Bool(active))
}
