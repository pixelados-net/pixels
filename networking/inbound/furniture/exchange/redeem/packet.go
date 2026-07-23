// Package redeem contains the ITEM_EXCHANGE_REDEEM inbound packet.
package redeem

import "github.com/niflaot/pixels/networking/codec"

// Header identifies ITEM_EXCHANGE_REDEEM.
const Header uint16 = 3115

// Payload stores one credit furniture redemption request.
type Payload struct {
	// ItemID identifies the owned placed or inventory furniture instance.
	ItemID int64
}

// Decode reads one credit furniture redemption request.
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
	return Payload{ItemID: int64(values[0].Int32)}, nil
}
