// Package completed encodes QUEST_COMPLETED responses.
package completed

import (
	"github.com/niflaot/pixels/networking/codec"
	questdata "github.com/niflaot/pixels/networking/outbound/progression/quest/data"
)

// Header identifies QUEST_COMPLETED.
const Header uint16 = 949

// Encode creates one quest completion response.
func Encode(quest questdata.Quest, showDialog bool) (codec.Packet, error) {
	payload, err := questdata.Append(nil, quest)
	if err == nil {
		payload, err = codec.AppendPayload(payload, codec.Definition{codec.BooleanField}, codec.Bool(showDialog))
	}
	return codec.Packet{Header: Header, Payload: payload}, err
}
