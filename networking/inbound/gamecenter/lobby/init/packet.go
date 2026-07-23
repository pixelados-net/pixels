// Package init decodes GAMES_INIT requests.
package init

import (
	"github.com/niflaot/pixels/networking/codec"
	"github.com/niflaot/pixels/networking/inbound"
)

// Header identifies GAMES_INIT.
const Header uint16 = 2914

// Decode returns the requested games init value.
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
