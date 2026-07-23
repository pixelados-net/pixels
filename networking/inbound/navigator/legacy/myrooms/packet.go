// Package myrooms decodes MY_ROOMS_SEARCH requests.
package myrooms

import (
	"github.com/niflaot/pixels/networking/codec"
	"github.com/niflaot/pixels/networking/inbound"
)

// Header identifies MY_ROOMS_SEARCH.
const Header uint16 = 2277

// Decode validates one header-only owned rooms request.
func Decode(packet codec.Packet) error {
	if err := inbound.ValidateHeader(packet, Header); err != nil {
		return err
	}
	_, err := codec.DecodePacketExact(packet, codec.Definition{})
	return err
}
