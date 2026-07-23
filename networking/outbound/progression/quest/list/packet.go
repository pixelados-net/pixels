// Package list encodes QUESTS_LIST responses.
package list

import (
	"github.com/niflaot/pixels/networking/codec"
	questdata "github.com/niflaot/pixels/networking/outbound/progression/quest/data"
)

// Header identifies QUESTS_LIST.
const Header uint16 = 3625

// Encode creates one quest catalog response.
func Encode(quests []questdata.Quest, openWindow bool) (codec.Packet, error) {
	payload, err := codec.AppendPayload(nil, codec.Definition{codec.Int32Field}, codec.Int32(int32(len(quests))))
	for _, quest := range quests {
		if err != nil {
			return codec.Packet{}, err
		}
		payload, err = questdata.Append(payload, quest)
	}
	if err == nil {
		payload, err = codec.AppendPayload(payload, codec.Definition{codec.BooleanField}, codec.Bool(openWindow))
	}
	return codec.Packet{Header: Header, Payload: payload}, err
}
