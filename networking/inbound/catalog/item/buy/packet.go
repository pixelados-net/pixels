// Package buy contains the PURCHASE_FROM_CATALOG inbound packet.
package buy

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the PURCHASE_FROM_CATALOG packet identifier.
	Header uint16 = 3492
)

// Payload contains the unpacked PURCHASE_FROM_CATALOG fields.
type Payload struct {
	// PageID identifies the source catalog page.
	PageID int32
	// OfferID identifies the purchased offer.
	OfferID int32
	// ExtraData stores client-supplied product data.
	ExtraData string
	// Amount stores the requested purchase quantity.
	Amount int32
}

// Definition describes Nitro's PurchaseFromCatalogComposer field order.
var Definition = codec.Definition{
	codec.Named("pageId", codec.Int32Field),
	codec.Named("offerId", codec.Int32Field),
	codec.Named("extraData", codec.StringField),
	codec.Named("amount", codec.Int32Field),
}

// Decode unpacks a PURCHASE_FROM_CATALOG packet payload.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return Payload{}, err
	}

	return Payload{PageID: values[0].Int32, OfferID: values[1].Int32, ExtraData: values[2].String, Amount: values[3].Int32}, nil
}
