// Package submittable decodes FORWARD_TO_A_SUBMITTABLE_ROOM requests.
package submittable

import (
	"github.com/niflaot/pixels/networking/codec"
	"github.com/niflaot/pixels/networking/inbound"
)

// Header identifies FORWARD_TO_A_SUBMITTABLE_ROOM.
const Header uint16 = 1450

// Decode validates one header-only request.
func Decode(packet codec.Packet) error {
	if err := inbound.ValidateHeader(packet, Header); err != nil {
		return err
	}
	_, err := codec.DecodePacketExact(packet, codec.Definition{})
	return err
}
