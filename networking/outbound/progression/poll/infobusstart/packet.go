// Package infobusstart encodes POLL_START_ROOM responses consumed by Nitro.
package infobusstart

import "github.com/niflaot/pixels/networking/codec"

// Header identifies POLL_START_ROOM.
const Header uint16 = 5200

// Encode creates one quick room poll with visible choices.
func Encode(question string, choices []string) (codec.Packet, error) {
	payload, err := codec.AppendPayload(nil, codec.Definition{codec.StringField, codec.Int32Field}, codec.String(question), codec.Int32(int32(len(choices))))
	for _, choice := range choices {
		if err != nil {
			return codec.Packet{}, err
		}
		payload, err = codec.AppendPayload(payload, codec.Definition{codec.StringField}, codec.String(choice))
	}
	return codec.Packet{Header: Header, Payload: payload}, err
}
