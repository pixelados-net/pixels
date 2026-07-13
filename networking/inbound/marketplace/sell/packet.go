// Package sell contains the SELL_MARKETPLACE_ITEM inbound packet.
package sell

import "github.com/niflaot/pixels/networking/codec"

// Header identifies SELL_MARKETPLACE_ITEM.
const Header uint16 = 3447

// Payload stores seller price, furniture type, and item id.
type Payload struct {
	// Price stores requested seller proceeds.
	Price int32
	// FurnitureType stores Nitro's floor or wall discriminator.
	FurnitureType int32
	// ItemID identifies the furniture instance.
	ItemID int32
}

// Decode reads a listing request.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, codec.Definition{codec.Named("price", codec.Int32Field), codec.Named("furnitureType", codec.Int32Field), codec.Named("itemId", codec.Int32Field)})
	if err != nil {
		return Payload{}, err
	}
	return Payload{Price: values[0].Int32, FurnitureType: values[1].Int32, ItemID: values[2].Int32}, nil
}
