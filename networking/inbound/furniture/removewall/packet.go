// Package removewall decodes REMOVE_WALL_ITEM requests.
package removewall

import (
	"github.com/niflaot/pixels/networking/codec"
	"github.com/niflaot/pixels/networking/inbound"
)

// Header identifies REMOVE_WALL_ITEM.
const Header uint16 = 3336

// Definition describes the wall item id.
var Definition = codec.Definition{codec.Named("itemId", codec.Int32Field)}

// Decode returns the wall item id.
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
