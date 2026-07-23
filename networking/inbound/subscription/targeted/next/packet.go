// Package next contains the GET_NEXT_TARGETED_OFFER inbound packet.
package next

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header identifies GET_NEXT_TARGETED_OFFER.
	Header uint16 = 596
)

// Payload contains the current targeted offer id.
type Payload struct {
	// OfferID identifies the offer to skip.
	OfferID int32
}

// Definition describes the packet payload.
var Definition = codec.Definition{codec.Int32Field}

// Decode decodes a GET_NEXT_TARGETED_OFFER packet.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	v, e := codec.DecodePacketExact(packet, Definition)
	if e != nil {
		return Payload{}, e
	}
	return Payload{OfferID: v[0].Int32}, nil
}
