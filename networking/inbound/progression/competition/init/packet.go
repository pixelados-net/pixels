// Package init decodes ROOM_COMPETITION_INIT requests.
package init

import (
	"github.com/niflaot/pixels/networking/codec"
	"github.com/niflaot/pixels/networking/inbound"
)

// Header identifies ROOM_COMPETITION_INIT.
const Header uint16 = 1334

// Decode validates one header-only request.
func Decode(packet codec.Packet) error {
	if err := inbound.ValidateHeader(packet, Header); err != nil {
		return err
	}
	_, err := codec.DecodePacketExact(packet, codec.Definition{})
	return err
}
