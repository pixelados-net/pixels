// Package answered encodes QUESTION_ANSWERED responses.
package answered

import "github.com/niflaot/pixels/networking/codec"

// Header identifies QUESTION_ANSWERED.
const Header uint16 = 2589

// Count stores one answer and aggregate count.
type Count struct {
	// Value identifies the answer.
	Value string
	// Count stores the aggregate count.
	Count int32
}

// Encode creates one live poll answer response.
func Encode(userID int32, value string, counts []Count) (codec.Packet, error) {
	payload, err := codec.AppendPayload(nil, codec.Definition{codec.Int32Field, codec.StringField, codec.Int32Field}, codec.Int32(userID), codec.String(value), codec.Int32(int32(len(counts))))
	for _, count := range counts {
		if err != nil {
			return codec.Packet{}, err
		}
		payload, err = codec.AppendPayload(payload, codec.Definition{codec.StringField, codec.Int32Field}, codec.String(count.Value), codec.Int32(count.Count))
	}
	return codec.Packet{Header: Header, Payload: payload}, err
}
