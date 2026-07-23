// Package randomstate decodes FURNITURE_RANDOMSTATE requests.
package randomstate

import (
	"github.com/niflaot/pixels/networking/codec"
	"github.com/niflaot/pixels/networking/inbound"
)

// Header identifies FURNITURE_RANDOMSTATE.
const Header uint16 = 3617

// Payload contains one random-state interaction.
type Payload struct {
	// ItemID identifies the clicked item.
	ItemID int32
	// State stores the client interaction parameter.
	State int32
}

// Definition describes the random-state request fields.
var Definition = codec.Definition{codec.Named("itemId", codec.Int32Field), codec.Named("state", codec.Int32Field)}

// Decode returns one random-state request.
func Decode(packet codec.Packet) (Payload, error) {
	if err := inbound.ValidateHeader(packet, Header); err != nil {
		return Payload{}, err
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return Payload{}, err
	}
	return Payload{ItemID: values[0].Int32, State: values[1].Int32}, nil
}
