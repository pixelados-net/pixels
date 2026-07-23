// Package wall decodes BUILDERS_CLUB_PLACE_WALL_ITEM requests.
package wall

import (
	"github.com/niflaot/pixels/networking/codec"
	"github.com/niflaot/pixels/networking/inbound"
)

// Header identifies BUILDERS_CLUB_PLACE_WALL_ITEM.
const Header uint16 = 462

// Payload contains one atomic catalog purchase and wall placement.
type Payload struct {
	// PageID identifies the catalog page.
	PageID int32
	// OfferID identifies the catalog offer.
	OfferID int32
	// ExtraData stores product customization.
	ExtraData string
	// WallPosition stores the Nitro wall coordinate string.
	WallPosition string
}

// Definition describes a Builders Club wall placement.
var Definition = codec.Definition{codec.Named("pageId", codec.Int32Field), codec.Named("offerId", codec.Int32Field), codec.Named("extraData", codec.StringField), codec.Named("wallPosition", codec.StringField)}

// Decode returns one Builders Club wall placement.
func Decode(packet codec.Packet) (Payload, error) {
	if err := inbound.ValidateHeader(packet, Header); err != nil {
		return Payload{}, err
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return Payload{}, err
	}
	return Payload{PageID: values[0].Int32, OfferID: values[1].Int32, ExtraData: values[2].String, WallPosition: values[3].String}, nil
}
