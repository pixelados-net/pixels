// Package purchase contains the CATALOG_PURCHASE_GIFT inbound packet.
package purchase

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header identifies CATALOG_PURCHASE_GIFT.
	Header uint16 = 1411
)

// Payload contains catalog gift purchase fields.
type Payload struct {
	// PageID identifies the source catalog page.
	PageID int32
	// ItemID identifies the purchased catalog item.
	ItemID int32
	// ExtraData stores client-provided product state.
	ExtraData string
	// ReceiverName identifies the gift recipient.
	ReceiverName string
	// GiftMessage stores the attached message.
	GiftMessage string
	// SpriteID identifies the selected wrapping sprite.
	SpriteID int32
	// BoxID identifies the selected wrapping box.
	BoxID int32
	// RibbonID identifies the selected ribbon.
	RibbonID int32
	// ShowMyFace reports whether sender identity is visible.
	ShowMyFace bool
}

// Definition describes the payload field order.
var Definition = codec.Definition{codec.Int32Field, codec.Int32Field, codec.StringField, codec.StringField, codec.StringField, codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.BooleanField}

// Decode decodes a catalog gift purchase.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	v, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return Payload{}, err
	}
	return Payload{PageID: v[0].Int32, ItemID: v[1].Int32, ExtraData: v[2].String, ReceiverName: v[3].String, GiftMessage: v[4].String, SpriteID: v[5].Int32, BoxID: v[6].Int32, RibbonID: v[7].Int32, ShowMyFace: v[8].Boolean}, nil
}
