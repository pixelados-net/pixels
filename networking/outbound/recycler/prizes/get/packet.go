// Package get contains the compatibility RECYCLER_PRIZES outbound packet.
package get

import (
	craftrecord "github.com/niflaot/pixels/internal/realm/crafting/record"
	"github.com/niflaot/pixels/networking/codec"
)

// Header identifies RECYCLER_PRIZES.
const Header uint16 = 3164

// Encode creates one rarity and prize table packet.
func Encode(prizes []craftrecord.Prize, chances [6]int32) (codec.Packet, error) {
	tierCount := int32(0)
	for tier := int32(1); tier <= 5; tier++ {
		for _, prize := range prizes {
			if prize.Tier == tier {
				tierCount++
				break
			}
		}
	}
	payload, err := codec.AppendPayload(nil, codec.Definition{codec.Int32Field}, codec.Int32(tierCount))
	if err != nil {
		return codec.Packet{}, err
	}
	for tier := int32(5); tier >= 1; tier-- {
		count := int32(0)
		for _, prize := range prizes {
			if prize.Tier == tier {
				count++
			}
		}
		if count == 0 {
			continue
		}
		payload, err = codec.AppendPayload(payload, codec.Definition{codec.Int32Field, codec.Int32Field, codec.Int32Field}, codec.Int32(tier), codec.Int32(chances[tier]), codec.Int32(count))
		if err != nil {
			return codec.Packet{}, err
		}
		for _, prize := range prizes {
			if prize.Tier != tier {
				continue
			}
			payload, err = codec.AppendPayload(payload, codec.Definition{codec.StringField, codec.Int32Field, codec.StringField, codec.Int32Field}, codec.String(prize.RewardName), codec.Int32(1), codec.String(prize.TypeCode), codec.Int32(prize.SpriteID))
			if err != nil {
				return codec.Packet{}, err
			}
		}
	}
	return codec.Packet{Header: Header, Payload: payload}, nil
}
