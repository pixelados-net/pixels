// Package stats contains the GET_MARKETPLACE_ITEM_STATS inbound packet.
package stats

import "github.com/niflaot/pixels/networking/codec"

// Header identifies GET_MARKETPLACE_ITEM_STATS.
const Header uint16 = 3288

// Payload stores the unknown category and sprite id.
type Payload struct {
	// Category stores Nitro's furniture category discriminator.
	Category int32
	// SpriteID identifies the requested furniture definition visually.
	SpriteID int32
}

// Decode reads the item-stat request.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, codec.Definition{codec.Named("category", codec.Int32Field), codec.Named("spriteId", codec.Int32Field)})
	if err != nil {
		return Payload{}, err
	}
	return Payload{Category: values[0].Int32, SpriteID: values[1].Int32}, nil
}
