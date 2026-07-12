// Package deactivate contains the ITEM_DICE_CLOSE inbound packet.
package deactivate

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the ITEM_DICE_CLOSE packet identifier.
	Header uint16 = 1533
)

// Payload contains the unpacked ITEM_DICE_CLOSE fields.
type Payload struct {
	// ItemID identifies the room dice item.
	ItemID int32
}

// Definition describes the ITEM_DICE_CLOSE payload fields.
var Definition = codec.Definition{codec.Named("itemId", codec.Int32Field)}

// Decode unpacks an ITEM_DICE_CLOSE packet payload.
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
