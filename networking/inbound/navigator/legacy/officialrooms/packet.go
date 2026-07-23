// Package officialrooms decodes GET_OFFICIAL_ROOMS requests.
package officialrooms

import (
	"github.com/niflaot/pixels/networking/codec"
	"github.com/niflaot/pixels/networking/inbound"
)

// Header identifies GET_OFFICIAL_ROOMS.
const Header uint16 = 1229

// Definition describes the mode filter carried by the legacy composer.
var Definition = codec.Definition{codec.Named("mode", codec.Int32Field)}

// Decode returns the requested mode value.
func Decode(packet codec.Packet) (int32, error) {
	if err := inbound.ValidateHeader(packet, Header); err != nil {
		return 0, err
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return 0, err
	}
	return values[0].Int32, nil
}
