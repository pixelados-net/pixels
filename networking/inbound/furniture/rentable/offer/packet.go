// Package offer decodes RENTABLE_GET_RENT_OR_BUYOUT_OFFER requests.
package offer

import (
	"github.com/niflaot/pixels/networking/codec"
	"github.com/niflaot/pixels/networking/inbound"
)

// Header identifies RENTABLE_GET_RENT_OR_BUYOUT_OFFER.
const Header uint16 = 2518

// Payload contains one rentable product offer request.
type Payload struct {
	// Wall reports whether the product is wall furniture.
	Wall bool
	// ProductName stores the furniture product identifier.
	ProductName string
	// Buyout requests a purchase offer instead of an extension offer.
	Buyout bool
}

// Definition describes the rentable product request.
var Definition = codec.Definition{codec.Named("wall", codec.BooleanField), codec.Named("productName", codec.StringField), codec.Named("buyout", codec.BooleanField)}

// Decode returns one rentable product offer request.
func Decode(packet codec.Packet) (Payload, error) {
	if err := inbound.ValidateHeader(packet, Header); err != nil {
		return Payload{}, err
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return Payload{}, err
	}
	return Payload{Wall: values[0].Boolean, ProductName: values[1].String, Buyout: values[2].Boolean}, nil
}
