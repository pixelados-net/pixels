// Package friendsrooms decodes ROOMS_WHERE_MY_FRIENDS_ARE requests.
package friendsrooms

import (
	"github.com/niflaot/pixels/networking/codec"
	"github.com/niflaot/pixels/networking/inbound"
)

// Header identifies ROOMS_WHERE_MY_FRIENDS_ARE.
const Header uint16 = 1786

// Decode validates one header-only friends' current rooms request.
func Decode(packet codec.Packet) error {
	if err := inbound.ValidateHeader(packet, Header); err != nil {
		return err
	}
	_, err := codec.DecodePacketExact(packet, codec.Definition{})
	return err
}
