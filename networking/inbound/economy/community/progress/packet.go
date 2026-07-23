// Package progress decodes GET_COMMUNITY_GOAL_PROGRESS requests.
package progress

import (
	"github.com/niflaot/pixels/networking/codec"
	"github.com/niflaot/pixels/networking/inbound"
)

// Header identifies GET_COMMUNITY_GOAL_PROGRESS.
const Header uint16 = 1145

// Decode validates one header-only community goal progress request.
func Decode(packet codec.Packet) error {
	if err := inbound.ValidateHeader(packet, Header); err != nil {
		return err
	}
	_, err := codec.DecodePacketExact(packet, codec.Definition{})
	return err
}
