// Package reject decodes REJECT_QUEST requests.
package reject

import (
	"github.com/niflaot/pixels/networking/codec"
	"github.com/niflaot/pixels/networking/inbound"
)

// Header identifies REJECT_QUEST.
const Header uint16 = 2397

// Decode validates one header-only quest rejection.
func Decode(packet codec.Packet) error {
	if err := inbound.ValidateHeader(packet, Header); err != nil {
		return err
	}
	_, err := codec.DecodePacketExact(packet, codec.Definition{})
	return err
}
