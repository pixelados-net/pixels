// Package viewed decodes ROOM_AD_EVENT_TAB_VIEWED telemetry.
package viewed

import (
	"github.com/niflaot/pixels/networking/codec"
	"github.com/niflaot/pixels/networking/inbound"
)

// Header identifies ROOM_AD_EVENT_TAB_VIEWED.
const Header uint16 = 2668

// Definition describes the header-only telemetry packet.
var Definition = codec.Definition{}

// Decode validates one tab-view telemetry packet.
func Decode(packet codec.Packet) error {
	if err := inbound.ValidateHeader(packet, Header); err != nil {
		return err
	}
	_, err := codec.DecodePacketExact(packet, Definition)
	return err
}
