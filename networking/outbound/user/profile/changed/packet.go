// Package changed contains the EXTENDED_PROFILE_CHANGED outbound packet.
package changed

import "github.com/niflaot/pixels/networking/codec"

// Header identifies EXTENDED_PROFILE_CHANGED.
const Header uint16 = 876

// Definition describes EXTENDED_PROFILE_CHANGED fields.
var Definition = codec.Definition{codec.Named("playerId", codec.Int32Field)}

// Encode creates an EXTENDED_PROFILE_CHANGED packet.
func Encode(playerID int32) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.Int32(playerID))
}
