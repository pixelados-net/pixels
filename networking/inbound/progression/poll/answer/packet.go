// Package answer decodes POLL_ANSWER requests.
package answer

import (
	"github.com/niflaot/pixels/networking/codec"
	"github.com/niflaot/pixels/networking/inbound"
)

// Header identifies POLL_ANSWER.
const Header uint16 = 3505

// Request stores one bounded poll answer.
type Request struct {
	// PollID identifies the poll.
	PollID int32
	// QuestionID identifies the question.
	QuestionID int32
	// Values stores submitted values.
	Values []string
}

// Decode returns one bounded poll answer request.
func Decode(packet codec.Packet) (Request, error) {
	if err := inbound.ValidateHeader(packet, Header); err != nil {
		return Request{}, err
	}
	values, rest, err := codec.DecodePacket(packet, codec.Definition{codec.Int32Field, codec.Int32Field, codec.Int32Field})
	if err != nil {
		return Request{}, err
	}
	count := values[2].Int32
	if count < 0 || count > 20 {
		return Request{}, codec.ErrInvalidField
	}
	request := Request{PollID: values[0].Int32, QuestionID: values[1].Int32, Values: make([]string, count)}
	for index := range request.Values {
		decoded, next, decodeErr := codec.DecodePayload(nil, codec.Definition{codec.StringField}, rest)
		if decodeErr != nil {
			return Request{}, decodeErr
		}
		request.Values[index], rest = decoded[0].String, next
	}
	if len(rest) != 0 {
		return Request{}, codec.ErrUnexpectedPayload
	}
	return request, nil
}
