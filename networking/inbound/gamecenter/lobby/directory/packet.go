// Package directory decodes GAME2CHECKGAMEDIRECTORYSTATUSMESSAGE requests.
package directory

import (
	"github.com/niflaot/pixels/networking/codec"
	"github.com/niflaot/pixels/networking/inbound"
)

// Header identifies GAME2CHECKGAMEDIRECTORYSTATUSMESSAGE.
const Header uint16 = 3259

// Decode validates one header-only directory status request.
func Decode(packet codec.Packet) error {
	if err := inbound.ValidateHeader(packet, Header); err != nil {
		return err
	}
	_, err := codec.DecodePacketExact(packet, codec.Definition{})
	return err
}
