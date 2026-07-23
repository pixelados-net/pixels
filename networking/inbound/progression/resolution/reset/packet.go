// Package reset decodes RESET_RESOLUTION_ACHIEVEMENT requests.
package reset

import (
	"github.com/niflaot/pixels/networking/codec"
	"github.com/niflaot/pixels/networking/inbound"
)

// Header identifies RESET_RESOLUTION_ACHIEVEMENT.
const Header uint16 = 3144

// Decode returns the resolution furniture identifier.
func Decode(packet codec.Packet) (int32, error) {
	if err := inbound.ValidateHeader(packet, Header); err != nil {
		return 0, err
	}
	values, err := codec.DecodePacketExact(packet, codec.Definition{codec.Int32Field})
	if err != nil {
		return 0, err
	}
	return values[0].Int32, nil
}
