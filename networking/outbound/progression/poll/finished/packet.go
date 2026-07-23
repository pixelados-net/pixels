// Package finished encodes QUESTION_FINISHED responses.
package finished

import (
	"github.com/niflaot/pixels/networking/codec"
	answered "github.com/niflaot/pixels/networking/outbound/progression/poll/answered"
)

// Header identifies QUESTION_FINISHED.
const Header uint16 = 1066

// Encode creates one final poll aggregate response.
func Encode(questionID int32, counts []answered.Count) (codec.Packet, error) {
	payload, err := codec.AppendPayload(nil, codec.Definition{codec.Int32Field, codec.Int32Field}, codec.Int32(questionID), codec.Int32(int32(len(counts))))
	for _, count := range counts {
		if err != nil {
			return codec.Packet{}, err
		}
		payload, err = codec.AppendPayload(payload, codec.Definition{codec.StringField, codec.Int32Field}, codec.String(count.Value), codec.Int32(count.Count))
	}
	return codec.Packet{Header: Header, Payload: payload}, err
}
