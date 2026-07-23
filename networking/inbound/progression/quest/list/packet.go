// Package list decodes GET_QUESTS requests.
package list

import (
	"github.com/niflaot/pixels/networking/codec"
	"github.com/niflaot/pixels/networking/inbound"
)

// Header identifies GET_QUESTS.
const Header uint16 = 3333

// Decode validates one header-only quest list request.
func Decode(packet codec.Packet) error { return decodeEmpty(packet) }

// decodeEmpty validates the packet header and empty payload.
func decodeEmpty(packet codec.Packet) error {
	if err := inbound.ValidateHeader(packet, Header); err != nil {
		return err
	}
	_, err := codec.DecodePacketExact(packet, codec.Definition{})
	return err
}
