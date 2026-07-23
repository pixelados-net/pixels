// Package recommendedrooms decodes MY_RECOMMENDED_ROOMS requests.
package recommendedrooms

import (
	"github.com/niflaot/pixels/networking/codec"
	"github.com/niflaot/pixels/networking/inbound"
)

// Header identifies MY_RECOMMENDED_ROOMS.
const Header uint16 = 2537

// Decode validates one header-only recommended rooms request.
func Decode(packet codec.Packet) error {
	if err := inbound.ValidateHeader(packet, Header); err != nil {
		return err
	}
	_, err := codec.DecodePacketExact(packet, codec.Definition{})
	return err
}
