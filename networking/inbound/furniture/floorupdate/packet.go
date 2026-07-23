// Package floorupdate contains the MOVE_FLOOR_ITEM inbound packet, used for both moving and rotating.
package floorupdate

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the MOVE_FLOOR_ITEM packet identifier.
	Header uint16 = 248
)

// Payload contains the unpacked MOVE_FLOOR_ITEM fields.
type Payload struct {
	// ItemID identifies the placed furniture item to reposition.
	ItemID int32

	// X stores the destination floor tile x coordinate.
	X int32

	// Y stores the destination floor tile y coordinate.
	Y int32

	// Rotation stores the destination floor instance rotation.
	Rotation int32
}

// Definition describes the MOVE_FLOOR_ITEM payload fields.
var Definition = codec.Definition{
	codec.Named("itemId", codec.Int32Field),
	codec.Named("x", codec.Int32Field),
	codec.Named("y", codec.Int32Field),
	codec.Named("rotation", codec.Int32Field),
}

// Decode unpacks a MOVE_FLOOR_ITEM packet payload.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return Payload{}, err
	}

	return Payload{ItemID: values[0].Int32, X: values[1].Int32, Y: values[2].Int32, Rotation: values[3].Int32}, nil
}
