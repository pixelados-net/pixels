// Package respect contains the USER_RESPECT outbound packet.
package respect

import "github.com/niflaot/pixels/networking/codec"

// Header identifies USER_RESPECT.
const Header uint16 = 2815

// Definition describes USER_RESPECT fields.
var Definition = codec.Definition{codec.Named("targetPlayerId", codec.Int32Field), codec.Named("totalReceived", codec.Int32Field)}

// Encode creates a USER_RESPECT packet.
func Encode(targetPlayerID int32, totalReceived int32) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.Int32(targetPlayerID), codec.Int32(totalReceived))
}
