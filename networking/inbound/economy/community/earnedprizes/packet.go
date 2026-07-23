// Package earnedprizes decodes GET_COMMUNITY_GOAL_EARNED_PRIZES requests.
package earnedprizes

import (
	"github.com/niflaot/pixels/networking/codec"
	"github.com/niflaot/pixels/networking/inbound"
)

// Header identifies GET_COMMUNITY_GOAL_EARNED_PRIZES.
const Header uint16 = 2688

// Decode validates one header-only earned community prizes request.
func Decode(packet codec.Packet) error {
	if err := inbound.ValidateHeader(packet, Header); err != nil {
		return err
	}
	_, err := codec.DecodePacketExact(packet, codec.Definition{})
	return err
}
