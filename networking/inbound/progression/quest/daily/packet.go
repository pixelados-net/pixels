// Package daily decodes GET_DAILY_QUEST requests.
package daily

import (
	"github.com/niflaot/pixels/networking/codec"
	"github.com/niflaot/pixels/networking/inbound"
)

// Header identifies GET_DAILY_QUEST.
const Header uint16 = 2486

// Request stores daily quest selection hints.
type Request struct {
	// Easy reports the requested difficulty.
	Easy bool
	// Index stores the client pool index.
	Index int32
}

// Decode returns one daily quest request.
func Decode(packet codec.Packet) (Request, error) {
	if err := inbound.ValidateHeader(packet, Header); err != nil {
		return Request{}, err
	}
	values, err := codec.DecodePacketExact(packet, codec.Definition{codec.BooleanField, codec.Int32Field})
	if err != nil {
		return Request{}, err
	}
	return Request{Easy: values[0].Boolean, Index: values[1].Int32}, nil
}
