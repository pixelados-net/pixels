// Package pickup contains the PICKUP_FLOOR_ITEM inbound packet.
package pickup

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the PICKUP_FLOOR_ITEM packet identifier.
	Header uint16 = 3456

	// FloorCategory identifies a floor item pickup request.
	FloorCategory int32 = 10

	// WallCategory identifies a wall item pickup request.
	WallCategory int32 = 20
)

// Payload contains the unpacked PICKUP_FLOOR_ITEM fields.
type Payload struct {
	// Category identifies whether the target is a floor or wall item.
	Category int32

	// ItemID identifies the placed furniture item to pick up.
	ItemID int32
}

// Definition describes the PICKUP_FLOOR_ITEM payload fields.
var Definition = codec.Definition{
	codec.Named("category", codec.Int32Field),
	codec.Named("itemId", codec.Int32Field),
}

// Decode unpacks a PICKUP_FLOOR_ITEM packet payload.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return Payload{}, err
	}

	return Payload{Category: values[0].Int32, ItemID: values[1].Int32}, nil
}
