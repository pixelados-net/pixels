// Package favouriterooms decodes MY_FAVOURITE_ROOMS_SEARCH requests.
package favouriterooms

import (
	"github.com/niflaot/pixels/networking/codec"
	"github.com/niflaot/pixels/networking/inbound"
)

// Header identifies MY_FAVOURITE_ROOMS_SEARCH.
const Header uint16 = 2578

// Decode validates one header-only favourite rooms request.
func Decode(packet codec.Packet) error {
	if err := inbound.ValidateHeader(packet, Header); err != nil {
		return err
	}
	_, err := codec.DecodePacketExact(packet, codec.Definition{})
	return err
}
