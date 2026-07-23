// Package results encodes QUIZ_RESULTS responses.
package results

import "github.com/niflaot/pixels/networking/codec"

// Header identifies QUIZ_RESULTS.
const Header uint16 = 2772

// Encode creates one quiz failure identifier response.
func Encode(code string, failedQuestionIDs []int32) (codec.Packet, error) {
	payload, err := codec.AppendPayload(nil, codec.Definition{codec.StringField, codec.Int32Field}, codec.String(code), codec.Int32(int32(len(failedQuestionIDs))))
	for _, id := range failedQuestionIDs {
		if err != nil {
			return codec.Packet{}, err
		}
		payload, err = codec.AppendPayload(payload, codec.Definition{codec.Int32Field}, codec.Int32(id))
	}
	return codec.Packet{Header: Header, Payload: payload}, err
}
