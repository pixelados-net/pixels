// Package submit decodes SUBMIT_ROOM_TO_COMPETITION requests.
package submit

import (
	"github.com/niflaot/pixels/networking/codec"
	"github.com/niflaot/pixels/networking/inbound"
)

// Header identifies SUBMIT_ROOM_TO_COMPETITION.
const Header uint16 = 2595

// Request stores one competition submission attempt.
type Request struct {
	// GoalCode identifies the competition.
	GoalCode string
	// ConfirmLevel stores the client confirmation phase.
	ConfirmLevel int32
}

// Decode returns one competition submission attempt.
func Decode(packet codec.Packet) (Request, error) {
	if err := inbound.ValidateHeader(packet, Header); err != nil {
		return Request{}, err
	}
	values, err := codec.DecodePacketExact(packet, codec.Definition{codec.StringField, codec.Int32Field})
	if err != nil {
		return Request{}, err
	}
	return Request{GoalCode: values[0].String, ConfirmLevel: values[1].Int32}, nil
}
