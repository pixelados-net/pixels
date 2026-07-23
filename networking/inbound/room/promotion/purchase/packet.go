// Package purchase decodes PURCHASE_ROOM_AD requests.
package purchase

import (
	"github.com/niflaot/pixels/networking/codec"
	"github.com/niflaot/pixels/networking/inbound"
)

// Header identifies PURCHASE_ROOM_AD.
const Header uint16 = 777

// Payload contains one room-promotion purchase.
type Payload struct {
	PageID      int32
	OfferID     int32
	RoomID      int32
	Title       string
	Extended    bool
	Description string
	CategoryID  int32
}

// Definition describes the room-ad purchase fields.
var Definition = codec.Definition{codec.Named("pageId", codec.Int32Field), codec.Named("offerId", codec.Int32Field), codec.Named("roomId", codec.Int32Field), codec.Named("title", codec.StringField), codec.Named("extended", codec.BooleanField), codec.Named("description", codec.StringField), codec.Named("categoryId", codec.Int32Field)}

// Decode returns one room-promotion purchase.
func Decode(packet codec.Packet) (Payload, error) {
	if err := inbound.ValidateHeader(packet, Header); err != nil {
		return Payload{}, err
	}
	v, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return Payload{}, err
	}
	return Payload{PageID: v[0].Int32, OfferID: v[1].Int32, RoomID: v[2].Int32, Title: v[3].String, Extended: v[4].Boolean, Description: v[5].String, CategoryID: v[6].Int32}, nil
}
