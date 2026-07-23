// Package accept decodes ACCEPT_QUEST requests.
package accept

import (
	"github.com/niflaot/pixels/networking/codec"
	"github.com/niflaot/pixels/networking/inbound"
)

// Header identifies ACCEPT_QUEST.
const Header uint16 = 3604

// Decode returns the accepted quest identifier.
func Decode(packet codec.Packet) (int32, error) {
	if err := inbound.ValidateHeader(packet, Header); err != nil {
		return 0, err
	}
	values, err := codec.DecodePacketExact(packet, codec.Definition{codec.Int32Field})
	if err != nil {
		return 0, err
	}
	return values[0].Int32, nil
}
