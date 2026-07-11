// Package request contains the ROOM_SETTINGS inbound packet.
package request

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header identifies ROOM_SETTINGS.
	Header uint16 = 3129
)

// Payload contains unpacked room settings request fields.
type Payload struct {
	// RoomID identifies the requested room.
	RoomID int32
}

// Definition describes ROOM_SETTINGS fields.
var Definition = codec.Definition{codec.Named("roomId", codec.Int32Field)}

// Decode unpacks a ROOM_SETTINGS packet.
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
