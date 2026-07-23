// Package confirm decodes FRIEND_FURNI_CONFIRM_LOCK requests.
package confirm

import (
	"github.com/niflaot/pixels/networking/codec"
	"github.com/niflaot/pixels/networking/inbound"
)

// Header identifies FRIEND_FURNI_CONFIRM_LOCK.
const Header uint16 = 3775

// Payload contains the pending lock decision.
type Payload struct {
	// ItemID identifies the lovelock furniture instance.
	ItemID int32
	// Confirmed reports whether the second player accepted.
	Confirmed bool
}

// Definition describes the renderer composer exactly.
var Definition = codec.Definition{codec.Named("itemId", codec.Int32Field), codec.Named("confirmed", codec.BooleanField)}

// Decode returns one pending lovelock decision.
func Decode(packet codec.Packet) (Payload, error) {
	if err := inbound.ValidateHeader(packet, Header); err != nil {
		return Payload{}, err
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return Payload{}, err
	}
	return Payload{ItemID: values[0].Int32, Confirmed: values[1].Boolean}, nil
}
