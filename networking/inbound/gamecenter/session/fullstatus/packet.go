// Package fullstatus decodes GAME2REQUESTFULLSTATUSUPDATEMESSAGE requests.
package fullstatus

import (
	"github.com/niflaot/pixels/networking/codec"
	"github.com/niflaot/pixels/networking/inbound"
)

// Header identifies GAME2REQUESTFULLSTATUSUPDATEMESSAGE.
const Header uint16 = 1598

// Decode returns the requested full game status value.
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
