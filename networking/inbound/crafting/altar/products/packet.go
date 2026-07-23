// Package products contains the functional GET_CRAFTABLE_PRODUCTS inbound packet.
package products

import "github.com/niflaot/pixels/networking/codec"

// Header identifies the integer-bearing altar request despite its crossed legacy name.
const Header uint16 = 1173

// Payload stores one altar inventory item request.
type Payload struct {
	// AltarItemID identifies the placed altar instance.
	AltarItemID int64
}

// Decode reads one altar inventory item request.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, codec.Definition{codec.Int32Field})
	if err != nil {
		return Payload{}, err
	}
	if values[0].Int32 <= 0 {
		return Payload{}, codec.ErrInvalidField
	}
	return Payload{AltarItemID: int64(values[0].Int32)}, nil
}
