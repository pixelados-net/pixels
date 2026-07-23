// Package friendsownedrooms decodes MY_FRIENDS_ROOM_SEARCH requests.
package friendsownedrooms

import (
	"github.com/niflaot/pixels/networking/codec"
	"github.com/niflaot/pixels/networking/inbound"
)

// Header identifies MY_FRIENDS_ROOM_SEARCH.
const Header uint16 = 2266

// Decode validates one header-only friend-owned rooms request.
func Decode(packet codec.Packet) error {
	if err := inbound.ValidateHeader(packet, Header); err != nil {
		return err
	}
	_, err := codec.DecodePacketExact(packet, codec.Definition{})
	return err
}
