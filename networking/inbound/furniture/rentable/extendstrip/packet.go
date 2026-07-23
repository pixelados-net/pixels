// Package extendstrip decodes RENTABLE_EXTEND_RENT_OR_BUYOUT_STRIP_ITEM requests.
package extendstrip

import (
	"github.com/niflaot/pixels/networking/codec"
	"github.com/niflaot/pixels/networking/inbound"
)

// Header identifies RENTABLE_EXTEND_RENT_OR_BUYOUT_STRIP_ITEM.
const Header uint16 = 2115

// Payload contains an inventory-rental mutation.
type Payload struct {
	// ItemID identifies the inventory furniture instance.
	ItemID int32
	// Buyout requests ownership transfer instead of extension.
	Buyout bool
}

// Definition describes the inventory-rental mutation.
var Definition = codec.Definition{codec.Named("itemId", codec.Int32Field), codec.Named("buyout", codec.BooleanField)}

// Decode returns one inventory-rental mutation.
func Decode(packet codec.Packet) (Payload, error) {
	if err := inbound.ValidateHeader(packet, Header); err != nil {
		return Payload{}, err
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return Payload{}, err
	}
	return Payload{ItemID: values[0].Int32, Buyout: values[1].Boolean}, nil
}
