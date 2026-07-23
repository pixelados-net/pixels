// Package cancel decodes CANCEL_QUEST requests.
package cancel

import (
	"github.com/niflaot/pixels/networking/codec"
	"github.com/niflaot/pixels/networking/inbound"
)

// Header identifies CANCEL_QUEST.
const Header uint16 = 3133

// Decode validates one header-only quest cancellation.
func Decode(packet codec.Packet) error {
	if err := inbound.ValidateHeader(packet, Header); err != nil {
		return err
	}
	_, err := codec.DecodePacketExact(packet, codec.Definition{})
	return err
}
