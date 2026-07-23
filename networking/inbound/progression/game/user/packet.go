// Package user decodes GET_USER_GAME_ACHIEVEMENTS requests.
package user

import (
	"github.com/niflaot/pixels/networking/codec"
	"github.com/niflaot/pixels/networking/inbound"
)

// Header identifies GET_USER_GAME_ACHIEVEMENTS.
const Header uint16 = 389

// Decode returns the requested player identifier.
func Decode(packet codec.Packet) (int32, error) {
	if err := inbound.ValidateHeader(packet, Header); err != nil {
		return 0, err
	}
	values, err := codec.DecodePacketExact(packet, codec.Definition{codec.Int32Field})
	if err != nil {
		return 0, err
	}
	return values[0].Int32, nil
}
