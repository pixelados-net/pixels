// Package seasonal encodes QUEST_SEASONAL responses.
package seasonal

import (
	"github.com/niflaot/pixels/networking/codec"
	questdata "github.com/niflaot/pixels/networking/outbound/progression/quest/data"
)

// Header identifies QUEST_SEASONAL.
const Header uint16 = 1122

// Encode creates one seasonal quest response.
func Encode(quests []questdata.Quest) (codec.Packet, error) {
	payload, err := codec.AppendPayload(nil, codec.Definition{codec.Int32Field}, codec.Int32(int32(len(quests))))
	for _, quest := range quests {
		if err != nil {
			return codec.Packet{}, err
		}
		payload, err = questdata.Append(payload, quest)
	}
	return codec.Packet{Header: Header, Payload: payload}, err
}
