// Package activate contains the ITEM_DICE_CLICK inbound packet.
package activate

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the ITEM_DICE_CLICK packet identifier.
	Header uint16 = 1990
)

// Payload contains the unpacked ITEM_DICE_CLICK fields.
type Payload struct {
	// ItemID identifies the room dice item.
	ItemID int32
}

// Definition describes the ITEM_DICE_CLICK payload fields.
var Definition = codec.Definition{codec.Named("itemId", codec.Int32Field)}

// Decode unpacks an ITEM_DICE_CLICK packet payload.
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
