// Package favoriteremove contains the ROOM_FAVORITE_REMOVE inbound packet.
package favoriteremove

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the ROOM_FAVORITE_REMOVE packet identifier.
	Header uint16 = 309
)

// Payload contains the unpacked ROOM_FAVORITE_REMOVE fields.
type Payload struct {
	// RoomID identifies the room to remove.
	RoomID int32
}

// Definition describes the ROOM_FAVORITE_REMOVE payload fields.
var Definition = codec.Definition{codec.Named("roomId", codec.Int32Field)}

// Decode unpacks a ROOM_FAVORITE_REMOVE packet payload.
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
