// Package onewaydoor decodes ONE_WAY_DOOR_CLICK requests.
package onewaydoor

import (
	"github.com/niflaot/pixels/networking/codec"
	"github.com/niflaot/pixels/networking/inbound"
)

// Header identifies ONE_WAY_DOOR_CLICK.
const Header uint16 = 2765

// Definition describes the clicked item.
var Definition = codec.Definition{codec.Named("itemId", codec.Int32Field)}

// Decode returns the clicked furniture item id.
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
