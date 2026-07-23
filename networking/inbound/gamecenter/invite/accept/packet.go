// Package accept decodes ACCEPTGAMEINVITE requests.
package accept

import (
	"github.com/niflaot/pixels/networking/codec"
	"github.com/niflaot/pixels/networking/inbound"
)

// Header identifies ACCEPTGAMEINVITE.
const Header uint16 = 3802

// Payload describes one game invite request.
type Payload struct {
	// GameTypeID stores the protocol gametypeid value.
	GameTypeID int32
	// InviterID stores the protocol inviterid value.
	InviterID int32
}

// Decode returns one validated game invite request.
func Decode(packet codec.Packet) (Payload, error) {
	if err := inbound.ValidateHeader(packet, Header); err != nil {
		return Payload{}, err
	}
	values, err := codec.DecodePacketExact(packet, codec.Definition{codec.Int32Field, codec.Int32Field})
	if err != nil {
		return Payload{}, err
	}
	return Payload{GameTypeID: values[0].Int32, InviterID: values[1].Int32}, nil
}
