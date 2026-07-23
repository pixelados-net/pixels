// Package list decodes GET_GAME_ACHIEVEMENTS requests.
package list

import (
	"github.com/niflaot/pixels/networking/codec"
	"github.com/niflaot/pixels/networking/inbound"
)

// Header identifies GET_GAME_ACHIEVEMENTS.
const Header uint16 = 2399

// Decode validates one header-only request.
func Decode(packet codec.Packet) error {
	if err := inbound.ValidateHeader(packet, Header); err != nil {
		return err
	}
	_, err := codec.DecodePacketExact(packet, codec.Definition{})
	return err
}
