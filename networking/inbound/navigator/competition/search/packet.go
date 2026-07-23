// Package search decodes COMPETITION_ROOM_SEARCH requests.
package search

import (
	"github.com/niflaot/pixels/networking/codec"
	"github.com/niflaot/pixels/networking/inbound"
)

// Header identifies COMPETITION_ROOM_SEARCH.
const Header uint16 = 433

// Request stores one paged competition room search.
type Request struct {
	// GoalID identifies the competition.
	GoalID int32
	// PageIndex stores the requested page.
	PageIndex int32
}

// Decode returns one paged competition room search.
func Decode(packet codec.Packet) (Request, error) {
	if err := inbound.ValidateHeader(packet, Header); err != nil {
		return Request{}, err
	}
	values, err := codec.DecodePacketExact(packet, codec.Definition{codec.Int32Field, codec.Int32Field})
	if err != nil {
		return Request{}, err
	}
	return Request{GoalID: values[0].Int32, PageIndex: values[1].Int32}, nil
}
