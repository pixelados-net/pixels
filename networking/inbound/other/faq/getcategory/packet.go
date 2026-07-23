// Package getcategory decodes the retired GET_FAQ_CATEGORY request.
package getcategory

import (
	"github.com/niflaot/pixels/networking/codec"
	"github.com/niflaot/pixels/networking/inbound"
)

// Header identifies GET_FAQ_CATEGORY.
const Header uint16 = 3445

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
