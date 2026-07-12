// Package friends contains the MESSENGER_FRIENDS outbound packet.
package friends

import (
	"github.com/niflaot/pixels/networking/codec"
	friendcard "github.com/niflaot/pixels/networking/outbound/messenger/friend/card"
)

// Header identifies MESSENGER_FRIENDS.
const Header uint16 = 3130

// Encode creates one friend-list fragment.
func Encode(totalFragments int32, fragmentNumber int32, cards []friendcard.Card) (codec.Packet, error) {
	payload, err := codec.AppendPayload(nil, codec.Definition{
		codec.Int32Field, codec.Int32Field, codec.Int32Field,
	}, codec.Int32(totalFragments), codec.Int32(fragmentNumber), codec.Int32(int32(len(cards))))
	if err != nil {
		return codec.Packet{}, err
	}
	for _, card := range cards {
		payload, err = friendcard.Append(payload, card)
		if err != nil {
			return codec.Packet{}, err
		}
	}
	return codec.Packet{Header: Header, Payload: payload}, nil
}
