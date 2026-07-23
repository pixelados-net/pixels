// Package vote decodes COMMUNITY_GOAL_VOTE_COMPOSER requests.
package vote

import (
	"github.com/niflaot/pixels/networking/codec"
	"github.com/niflaot/pixels/networking/inbound"
)

// Header identifies COMMUNITY_GOAL_VOTE_COMPOSER.
const Header uint16 = 3536

// Definition describes the selected community goal vote option.
var Definition = codec.Definition{codec.Named("voteOption", codec.Int32Field)}

// Decode returns the selected vote option.
func Decode(packet codec.Packet) (int32, error) {
	if err := inbound.ValidateHeader(packet, Header); err != nil {
		return 0, err
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return 0, err
	}
	return values[0].Int32, nil
}
