// Package items contains Nitro's inventory unseen-items acknowledgement.
package items

import (
	"encoding/binary"

	"github.com/niflaot/pixels/networking/codec"
)

// Header identifies UNSEEN_RESET_ITEMS.
const Header uint16 = 2343

// Payload contains one acknowledged inventory category and its item ids.
type Payload struct {
	// Category identifies the acknowledged inventory category.
	Category int32
	// ItemIDs identifies the acknowledged inventory entries.
	ItemIDs []int64
}

// Decode unpacks one bounded unseen-items acknowledgement.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, rest, err := codec.DecodePacket(packet, codec.Definition{codec.Int32Field, codec.Int32Field})
	if err != nil {
		return Payload{}, err
	}
	count := int(values[1].Int32)
	if count < 0 || count > len(rest)/4 || len(rest) != count*4 {
		return Payload{}, codec.ErrUnexpectedPayload
	}
	payload := Payload{Category: values[0].Int32, ItemIDs: make([]int64, count)}
	for index := range payload.ItemIDs {
		payload.ItemIDs[index] = int64(int32(binary.BigEndian.Uint32(rest[index*4:])))
	}
	return payload, nil
}
