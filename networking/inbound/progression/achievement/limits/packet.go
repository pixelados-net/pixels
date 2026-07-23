// Package limits decodes GET_BADGE_POINT_LIMITS requests.
package limits

import (
	"github.com/niflaot/pixels/networking/codec"
	"github.com/niflaot/pixels/networking/inbound"
)

// Header identifies GET_BADGE_POINT_LIMITS.
const Header uint16 = 1371

// Decode validates one header-only badge point limit request.
func Decode(packet codec.Packet) error {
	if err := inbound.ValidateHeader(packet, Header); err != nil {
		return err
	}
	_, err := codec.DecodePacketExact(packet, codec.Definition{})
	return err
}
