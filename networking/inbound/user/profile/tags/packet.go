// Package tags contains the USER_TAGS inbound packet.
package tags

import "github.com/niflaot/pixels/networking/codec"

// Header identifies USER_TAGS.
const Header uint16 = 17

// Definition describes USER_TAGS fields.
var Definition = codec.Definition{codec.Named("roomUnitId", codec.Int32Field)}

// Payload contains decoded USER_TAGS fields.
type Payload struct {
	// RoomUnitID identifies the room-local avatar unit.
	RoomUnitID int32
}

// Decode decodes USER_TAGS.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return Payload{}, err
	}
	return Payload{RoomUnitID: values[0].Int32}, nil
}
