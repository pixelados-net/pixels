// Package resetstate decodes the retired RESET_PHONE_NUMBER_STATE request.
package resetstate

import (
	"github.com/niflaot/pixels/networking/codec"
	"github.com/niflaot/pixels/networking/inbound"
)

// Header identifies RESET_PHONE_NUMBER_STATE.
const Header uint16 = 2741

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
