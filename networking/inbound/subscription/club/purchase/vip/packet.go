// Package vip contains the PURCHASE_VIP_MEMBERSHIP_EXTENSION inbound packet.
package vip

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the PURCHASE_VIP_MEMBERSHIP_EXTENSION packet identifier.
	Header uint16 = 3407
)

// Payload contains a selected VIP membership offer.
type Payload struct {
	// OfferID identifies the selected extension offer.
	OfferID int32
}

// Definition describes the required packet fields.
var Definition = codec.Definition{codec.Int32Field}

// Decode decodes a PURCHASE_VIP_MEMBERSHIP_EXTENSION packet.
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
