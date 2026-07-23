// Package seasonal decodes GET_SEASONAL_QUESTS requests.
package seasonal

import (
	"github.com/niflaot/pixels/networking/codec"
	"github.com/niflaot/pixels/networking/inbound"
)

// Header identifies GET_SEASONAL_QUESTS.
const Header uint16 = 1190

// Decode validates one header-only seasonal quest request.
func Decode(packet codec.Packet) error {
	if err := inbound.ValidateHeader(packet, Header); err != nil {
		return err
	}
	_, err := codec.DecodePacketExact(packet, codec.Definition{})
	return err
}
