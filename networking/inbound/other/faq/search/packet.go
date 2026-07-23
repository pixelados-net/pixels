// Package search decodes the retired SEARCH_FAQS request.
package search

import (
	"github.com/niflaot/pixels/networking/codec"
	"github.com/niflaot/pixels/networking/inbound"
)

// Header identifies SEARCH_FAQS.
const Header uint16 = 2031

// Definition describes the header-only request.
var Definition = codec.Definition{}

// Decode validates the retired request exactly.
func Decode(packet codec.Packet) error {
	if err := inbound.ValidateHeader(packet, Header); err != nil {
		return err
	}
	_, err := codec.DecodePacketExact(packet, Definition)
	return err
}
