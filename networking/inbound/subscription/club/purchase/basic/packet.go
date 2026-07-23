// Package basic contains the PURCHASE_BASIC_MEMBERSHIP_EXTENSION inbound packet.
package basic

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the PURCHASE_BASIC_MEMBERSHIP_EXTENSION packet identifier.
	Header uint16 = 2735
)

// Payload contains a selected basic membership offer.
type Payload struct {
	// OfferID identifies the selected extension offer.
	OfferID int32
}

// Definition describes the required packet fields.
var Definition = codec.Definition{codec.Int32Field}

// Decode decodes a PURCHASE_BASIC_MEMBERSHIP_EXTENSION packet.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return Payload{}, err
	}

	return Payload{OfferID: values[0].Int32}, nil
}
