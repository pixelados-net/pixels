// Package redeemprize decodes REDEEM_COMMUNITY_GOAL_PRIZE requests.
package redeemprize

import (
	"github.com/niflaot/pixels/networking/codec"
	"github.com/niflaot/pixels/networking/inbound"
)

// Header identifies REDEEM_COMMUNITY_GOAL_PRIZE.
const Header uint16 = 90

// Definition describes the requested community goal prize.
var Definition = codec.Definition{codec.Named("communityGoalId", codec.Int32Field)}

// Decode returns the requested community goal identifier.
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
