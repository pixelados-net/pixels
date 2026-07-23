// Package timing decodes GET_CURRENT_TIMING_CODE requests.
package timing

import (
	"github.com/niflaot/pixels/networking/codec"
	"github.com/niflaot/pixels/networking/inbound"
)

// Header identifies GET_CURRENT_TIMING_CODE.
const Header uint16 = 2912

// Decode validates one header-only timing request.
func Decode(packet codec.Packet) error {
	if err := inbound.ValidateHeader(packet, Header); err != nil {
		return err
	}
	_, err := codec.DecodePacketExact(packet, codec.Definition{})
	return err
}
