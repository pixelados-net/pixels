// Package roomrights decodes MY_ROOM_RIGHTS_SEARCH requests.
package roomrights

import (
	"github.com/niflaot/pixels/networking/codec"
	"github.com/niflaot/pixels/networking/inbound"
)

// Header identifies MY_ROOM_RIGHTS_SEARCH.
const Header uint16 = 272

// Decode validates one header-only room rights request.
func Decode(packet codec.Packet) error {
	if err := inbound.ValidateHeader(packet, Header); err != nil {
		return err
	}
	_, err := codec.DecodePacketExact(packet, codec.Definition{})
	return err
}
