// Package relinquish contains the ROOM_RIGHTS_REMOVE_OWN inbound packet.
package relinquish

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header identifies ROOM_RIGHTS_REMOVE_OWN.
	Header uint16 = 3182
)

// Payload contains unpacked relinquish fields.
type Payload struct {
	// RoomID identifies the room.
	RoomID int32
}

// Definition describes ROOM_RIGHTS_REMOVE_OWN fields.
var Definition = codec.Definition{codec.Named("roomId", codec.Int32Field)}

// Decode unpacks a ROOM_RIGHTS_REMOVE_OWN packet.
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
