// Package promoted decodes the retired FORWARD_TO_RANDOM_PROMOTED_ROOM request.
package promoted

import (
	"github.com/niflaot/pixels/networking/codec"
	"github.com/niflaot/pixels/networking/inbound"
)

// Header identifies FORWARD_TO_RANDOM_PROMOTED_ROOM.
const Header uint16 = 10

// Definition describes the legacy room-category code.
var Definition = codec.Definition{codec.Named("categoryCode", codec.StringField)}

// Decode returns the compatibility-only category code.
func Decode(packet codec.Packet) (string, error) {
	if err := inbound.ValidateHeader(packet, Header); err != nil {
		return "", err
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return "", err
	}
	return values[0].String, nil
}
