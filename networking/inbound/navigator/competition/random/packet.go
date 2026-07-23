// Package random decodes FORWARD_TO_RANDOM_COMPETITION_ROOM requests.
package random

import (
	"github.com/niflaot/pixels/networking/codec"
	"github.com/niflaot/pixels/networking/inbound"
)

// Header identifies FORWARD_TO_RANDOM_COMPETITION_ROOM.
const Header uint16 = 865

// Decode returns the requested competition goal code.
func Decode(packet codec.Packet) (string, error) {
	if err := inbound.ValidateHeader(packet, Header); err != nil {
		return "", err
	}
	values, err := codec.DecodePacketExact(packet, codec.Definition{codec.StringField})
	if err != nil {
		return "", err
	}
	return values[0].String, nil
}
