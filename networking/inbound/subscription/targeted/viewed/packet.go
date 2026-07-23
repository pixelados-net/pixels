// Package viewed contains the SHOP_TARGETED_OFFER_VIEWED inbound packet.
package viewed

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header identifies SHOP_TARGETED_OFFER_VIEWED.
	Header uint16 = 3483
)

// Payload contains a targeted offer view state.
type Payload struct {
	// OfferID identifies the targeted offer.
	OfferID int32
	// State stores the client view state.
	State int32
}

// Definition describes the packet payload.
var Definition = codec.Definition{codec.Int32Field, codec.Int32Field}

// Decode decodes a SHOP_TARGETED_OFFER_VIEWED packet.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	v, e := codec.DecodePacketExact(packet, Definition)
	if e != nil {
		return Payload{}, e
	}
	return Payload{OfferID: v[0].Int32, State: v[1].Int32}, nil
}
