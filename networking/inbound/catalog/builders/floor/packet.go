// Package floor decodes BUILDERS_CLUB_PLACE_ROOM_ITEM requests.
package floor

import (
	"github.com/niflaot/pixels/networking/codec"
	"github.com/niflaot/pixels/networking/inbound"
)

// Header identifies BUILDERS_CLUB_PLACE_ROOM_ITEM.
const Header uint16 = 1051

// Payload contains one atomic catalog purchase and floor placement.
type Payload struct {
	// PageID identifies the catalog page.
	PageID int32
	// OfferID identifies the catalog offer.
	OfferID int32
	// ExtraData stores product customization.
	ExtraData string
	// X stores the destination x coordinate.
	X int32
	// Y stores the destination y coordinate.
	Y int32
	// Direction stores the destination rotation.
	Direction int32
}

// Definition describes a Builders Club floor placement.
var Definition = codec.Definition{codec.Named("pageId", codec.Int32Field), codec.Named("offerId", codec.Int32Field), codec.Named("extraData", codec.StringField), codec.Named("x", codec.Int32Field), codec.Named("y", codec.Int32Field), codec.Named("direction", codec.Int32Field)}

// Decode returns one Builders Club floor placement.
func Decode(packet codec.Packet) (Payload, error) {
	if err := inbound.ValidateHeader(packet, Header); err != nil {
		return Payload{}, err
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return Payload{}, err
	}
	return Payload{PageID: values[0].Int32, OfferID: values[1].Int32, ExtraData: values[2].String, X: values[3].Int32, Y: values[4].Int32, Direction: values[5].Int32}, nil
}
