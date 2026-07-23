// Package result encodes the WIRED_REWARD outbound packet.
package result

import "github.com/niflaot/pixels/networking/codec"

// Header is the WIRED_REWARD packet identifier.
const Header uint16 = 178

// Definition describes WIRED_REWARD.
var Definition = codec.Definition{codec.Named("reason", codec.Int32Field)}

// Encode creates a WIRED reward-result packet.
func Encode(reason int32) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.Int32(reason))
}
