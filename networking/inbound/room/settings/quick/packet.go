// Package quick contains the UPDATE_ROOM_CATEGORY_AND_TRADE inbound packet.
package quick

import "github.com/niflaot/pixels/networking/codec"

// Header identifies UPDATE_ROOM_CATEGORY_AND_TRADE.
const Header uint16 = 1265

// Payload stores the focused room settings mutation.
type Payload struct {
	// RoomID identifies the room.
	RoomID int32
	// CategoryID identifies the navigator category.
	CategoryID int32
	// TradeMode stores the room direct-trade policy.
	TradeMode int32
}

// Decode reads the focused room settings mutation.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, codec.Definition{codec.Int32Field, codec.Int32Field, codec.Int32Field})
	if err != nil {
		return Payload{}, err
	}
	return Payload{RoomID: values[0].Int32, CategoryID: values[1].Int32, TradeMode: values[2].Int32}, nil
}
