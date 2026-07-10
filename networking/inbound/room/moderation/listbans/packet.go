// Package listbans contains the ROOM_BAN_LIST inbound packet.
package listbans

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header identifies ROOM_BAN_LIST.
	Header uint16 = 2267
)

// Payload contains unpacked ban-list fields.
type Payload struct {
	// RoomID identifies the room.
	RoomID int32
}

// Definition describes ROOM_BAN_LIST fields.
var Definition = codec.Definition{codec.Named("roomId", codec.Int32Field)}

// Decode unpacks a ROOM_BAN_LIST packet.
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
