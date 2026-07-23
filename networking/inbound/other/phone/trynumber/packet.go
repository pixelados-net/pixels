// Package trynumber decodes the retired TRY_PHONE_NUMBER request.
package trynumber

import (
	"github.com/niflaot/pixels/networking/codec"
	"github.com/niflaot/pixels/networking/inbound"
)

// Header identifies TRY_PHONE_NUMBER.
const Header uint16 = 790

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
