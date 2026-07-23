// Package offerrewarddelivered encodes the retired reward-delivery packet.
package offerrewarddelivered

import "github.com/niflaot/pixels/networking/codec"

// Header identifies NOTIFICATION_OFFER_REWARD_DELIVERED.
const Header uint16 = 2125

// Definition describes the header-only legacy packet.
var Definition = codec.Definition{}

// Encode creates one compatibility packet.
func Encode() (codec.Packet, error) {
	return codec.NewPacket(Header, Definition)
}
