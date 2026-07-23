// Package answers decodes POST_QUIZ_ANSWERS requests.
package answers

import (
	"github.com/niflaot/pixels/networking/codec"
	"github.com/niflaot/pixels/networking/inbound"
)

// Header identifies POST_QUIZ_ANSWERS.
const Header uint16 = 3720

// Request stores ordered quiz answer identifiers.
type Request struct {
	// Code identifies the quiz.
	Code string
	// AnswerIDs stores answers in question order.
	AnswerIDs []int32
}

// Decode returns one bounded quiz submission.
func Decode(packet codec.Packet) (Request, error) {
	if err := inbound.ValidateHeader(packet, Header); err != nil {
		return Request{}, err
	}
	values, rest, err := codec.DecodePacket(packet, codec.Definition{codec.StringField, codec.Int32Field})
	if err != nil {
		return Request{}, err
	}
	count := values[1].Int32
	if count < 0 || count > 100 {
		return Request{}, codec.ErrInvalidField
	}
	request := Request{Code: values[0].String, AnswerIDs: make([]int32, count)}
	for index := range request.AnswerIDs {
		decoded, next, decodeErr := codec.DecodePayload(nil, codec.Definition{codec.Int32Field}, rest)
		if decodeErr != nil {
			return Request{}, decodeErr
		}
		request.AnswerIDs[index], rest = decoded[0].Int32, next
	}
	if len(rest) != 0 {
		return Request{}, codec.ErrUnexpectedPayload
	}
	return request, nil
}
