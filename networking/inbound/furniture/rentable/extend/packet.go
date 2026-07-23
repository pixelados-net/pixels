// Package extend decodes RENTABLE_EXTEND_RENT_OR_BUYOUT_FURNI requests.
package extend

import (
	"github.com/niflaot/pixels/networking/codec"
	"github.com/niflaot/pixels/networking/inbound"
)

// Header identifies RENTABLE_EXTEND_RENT_OR_BUYOUT_FURNI.
const Header uint16 = 1071

// Payload contains a placed-furniture rent mutation.
type Payload struct {
	// Wall reports whether the target is wall furniture.
	Wall bool
	// ItemID identifies the room furniture instance.
	ItemID int32
	// Buyout requests ownership transfer instead of extension.
	Buyout bool
}

// Definition describes the rentable furniture mutation.
var Definition = codec.Definition{codec.Named("wall", codec.BooleanField), codec.Named("itemId", codec.Int32Field), codec.Named("buyout", codec.BooleanField)}

// Decode returns one rentable furniture mutation.
func Decode(packet codec.Packet) (Payload, error) {
	if err := inbound.ValidateHeader(packet, Header); err != nil {
		return Payload{}, err
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return Payload{}, err
	}
	return Payload{Wall: values[0].Boolean, ItemID: values[1].Int32, Buyout: values[2].Boolean}, nil
}
