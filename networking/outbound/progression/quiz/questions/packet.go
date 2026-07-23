// Package questions encodes QUIZ_DATA responses.
package questions

import "github.com/niflaot/pixels/networking/codec"

// Header identifies QUIZ_DATA.
const Header uint16 = 2927

// Encode creates one quiz question identifier response.
func Encode(code string, questionIDs []int32) (codec.Packet, error) {
	payload, err := codec.AppendPayload(nil, codec.Definition{codec.StringField, codec.Int32Field}, codec.String(code), codec.Int32(int32(len(questionIDs))))
	for _, id := range questionIDs {
		if err != nil {
			return codec.Packet{}, err
		}
		payload, err = codec.AppendPayload(payload, codec.Definition{codec.Int32Field}, codec.Int32(id))
	}
	return codec.Packet{Header: Header, Payload: payload}, err
}
