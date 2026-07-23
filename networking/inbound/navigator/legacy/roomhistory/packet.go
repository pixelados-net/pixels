// Package roomhistory decodes MY_ROOM_HISTORY_SEARCH requests.
package roomhistory

import (
	"github.com/niflaot/pixels/networking/codec"
	"github.com/niflaot/pixels/networking/inbound"
)

// Header identifies MY_ROOM_HISTORY_SEARCH.
const Header uint16 = 2264

// Decode validates one header-only room history request.
func Decode(packet codec.Packet) error {
	if err := inbound.ValidateHeader(packet, Header); err != nil {
		return err
	}
	_, err := codec.DecodePacketExact(packet, codec.Definition{})
	return err
}
