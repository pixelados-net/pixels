// Package exit decodes GAME2EXITGAMEMESSAGE requests.
package exit

import (
	"github.com/niflaot/pixels/networking/codec"
	"github.com/niflaot/pixels/networking/inbound"
)

// Header identifies GAME2EXITGAMEMESSAGE.
const Header uint16 = 1445

// Decode returns the requested game exit value.
func Decode(packet codec.Packet) (bool, error) {
	if err := inbound.ValidateHeader(packet, Header); err != nil {
		return false, err
	}
	values, err := codec.DecodePacketExact(packet, codec.Definition{codec.BooleanField})
	if err != nil {
		return false, err
	}
	return values[0].Boolean, nil
}
