// Package partof decodes GET_IS_USER_PART_OF_COMPETITION requests.
package partof

import (
	"github.com/niflaot/pixels/networking/codec"
	"github.com/niflaot/pixels/networking/inbound"
)

// Header identifies GET_IS_USER_PART_OF_COMPETITION.
const Header uint16 = 2077

// Decode returns the requested goal code.
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
