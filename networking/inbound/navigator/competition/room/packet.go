// Package room decodes FORWARD_TO_A_COMPETITION_ROOM requests.
package room

import (
	"github.com/niflaot/pixels/networking/codec"
	"github.com/niflaot/pixels/networking/inbound"
)

// Header identifies FORWARD_TO_A_COMPETITION_ROOM.
const Header uint16 = 172

// Request stores one competition room request.
type Request struct {
	// GoalCode identifies the competition.
	GoalCode string
	// Index stores the requested result index.
	Index int32
}

// Decode returns one competition room request.
func Decode(packet codec.Packet) (Request, error) {
	if err := inbound.ValidateHeader(packet, Header); err != nil {
		return Request{}, err
	}
	values, err := codec.DecodePacketExact(packet, codec.Definition{codec.StringField, codec.Int32Field})
	if err != nil {
		return Request{}, err
	}
	return Request{GoalCode: values[0].String, Index: values[1].Int32}, nil
}
