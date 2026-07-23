// Package vote decodes the retired VOTE_FOR_ROOM request.
package vote

import (
	"github.com/niflaot/pixels/networking/codec"
	"github.com/niflaot/pixels/networking/inbound"
)

// Header identifies VOTE_FOR_ROOM.
const Header uint16 = 143

// Definition describes the header-only request.
var Definition = codec.Definition{}

// Decode validates the retired request exactly.
func Decode(packet codec.Packet) error {
	if err := inbound.ValidateHeader(packet, Header); err != nil {
		return err
	}
	_, err := codec.DecodePacketExact(packet, Definition)
	return err
}
