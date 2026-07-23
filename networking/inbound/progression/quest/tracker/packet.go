// Package tracker decodes OPEN_QUEST_TRACKER requests.
package tracker

import (
	"github.com/niflaot/pixels/networking/codec"
	"github.com/niflaot/pixels/networking/inbound"
)

// Header identifies OPEN_QUEST_TRACKER.
const Header uint16 = 2750

// Decode validates one header-only quest tracker request.
func Decode(packet codec.Packet) error {
	if err := inbound.ValidateHeader(packet, Header); err != nil {
		return err
	}
	_, err := codec.DecodePacketExact(packet, codec.Definition{})
	return err
}
