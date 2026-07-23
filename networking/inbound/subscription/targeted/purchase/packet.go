// Package purchase contains the PURCHASE_TARGETED_OFFER inbound packet.
package purchase

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header identifies PURCHASE_TARGETED_OFFER.
	Header uint16 = 1826
)

// Payload contains a targeted offer purchase.
type Payload struct {
	// OfferID identifies the targeted offer.
	OfferID int32
	// Quantity stores requested units.
	Quantity int32
}

// Definition describes the packet payload.
var Definition = codec.Definition{codec.Int32Field, codec.Int32Field}

// Decode decodes a PURCHASE_TARGETED_OFFER packet.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	v, e := codec.DecodePacketExact(packet, Definition)
	if e != nil {
		return Payload{}, e
	}
	return Payload{OfferID: v[0].Int32, Quantity: v[1].Int32}, nil
}
