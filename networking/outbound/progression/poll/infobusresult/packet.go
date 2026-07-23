// Package infobusresult encodes POLL_ROOM_RESULT responses consumed by Nitro.
package infobusresult

import "github.com/niflaot/pixels/networking/codec"

// Header identifies POLL_ROOM_RESULT.
const Header uint16 = 5201

// Choice describes one quick-poll result.
type Choice struct {
	// Text stores the visible answer.
	Text string
	// Votes stores its vote count.
	Votes int32
}

// Encode creates one room poll result snapshot.
func Encode(question string, choices []Choice, totalVotes int32) (codec.Packet, error) {
	payload, err := codec.AppendPayload(nil, codec.Definition{codec.StringField, codec.Int32Field}, codec.String(question), codec.Int32(int32(len(choices))))
	for _, choice := range choices {
		if err != nil {
			return codec.Packet{}, err
		}
		payload, err = codec.AppendPayload(payload, codec.Definition{codec.StringField, codec.Int32Field}, codec.String(choice.Text), codec.Int32(choice.Votes))
	}
	if err != nil {
		return codec.Packet{}, err
	}
	payload, err = codec.AppendPayload(payload, codec.Definition{codec.Int32Field}, codec.Int32(totalVotes))
	return codec.Packet{Header: Header, Payload: payload}, err
}
