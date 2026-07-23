// Package colorwheel contains the ITEM_COLOR_WHEEL_CLICK inbound packet.
package colorwheel

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the ITEM_COLOR_WHEEL_CLICK packet identifier.
	Header uint16 = 2144
)

// Payload contains the unpacked ITEM_COLOR_WHEEL_CLICK fields.
type Payload struct {
	// ItemID identifies the room color-wheel item.
	ItemID int32
}

// Definition describes the ITEM_COLOR_WHEEL_CLICK payload fields.
var Definition = codec.Definition{codec.Named("itemId", codec.Int32Field)}

// Decode unpacks an ITEM_COLOR_WHEEL_CLICK packet payload.
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
