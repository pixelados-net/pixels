// Package homeroom contains the USER_HOME_ROOM inbound packet.
package homeroom

import "github.com/niflaot/pixels/networking/codec"

// Header identifies USER_HOME_ROOM.
const Header uint16 = 1740

// Definition describes USER_HOME_ROOM fields.
var Definition = codec.Definition{codec.Named("roomId", codec.Int32Field)}

// Payload contains decoded home-room settings.
type Payload struct {
	// RoomID identifies the requested home room or zero to clear it.
	RoomID int32
}

// Decode decodes USER_HOME_ROOM.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return Payload{}, err
	}
	return Payload{RoomID: values[0].Int32}, nil
}
