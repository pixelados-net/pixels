// Package daily encodes QUEST_DAILY responses.
package daily

import (
	"github.com/niflaot/pixels/networking/codec"
	questdata "github.com/niflaot/pixels/networking/outbound/progression/quest/data"
)

// Header identifies QUEST_DAILY.
const Header uint16 = 1878

// Encode creates one optional daily quest response.
func Encode(quest *questdata.Quest, easyCount int32, hardCount int32) (codec.Packet, error) {
	payload, err := codec.AppendPayload(nil, codec.Definition{codec.BooleanField}, codec.Bool(quest != nil))
	if quest != nil && err == nil {
		payload, err = questdata.Append(payload, *quest)
	}
	if quest != nil && err == nil {
		payload, err = codec.AppendPayload(payload, codec.Definition{codec.Int32Field, codec.Int32Field}, codec.Int32(easyCount), codec.Int32(hardCount))
	}
	return codec.Packet{Header: Header, Payload: payload}, err
}
