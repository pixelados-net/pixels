// Package redeem contains the ITEM_CLOTHING_REDEEM inbound packet.
package redeem

import "github.com/niflaot/pixels/networking/codec"

// Header identifies ITEM_CLOTHING_REDEEM.
const Header uint16 = 3374

// Definition describes ITEM_CLOTHING_REDEEM fields.
var Definition = codec.Definition{codec.Named("itemId", codec.Int32Field)}

// Payload contains one clothing furniture redemption request.
type Payload struct {
	// ItemID identifies an inventory furniture instance.
	ItemID int32
}

// Decode decodes ITEM_CLOTHING_REDEEM.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return Payload{}, err
	}
	return Payload{ItemID: values[0].Int32}, nil
}
