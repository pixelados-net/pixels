// Package favoriteadd contains the ROOM_FAVORITE inbound packet.
package favoriteadd

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the ROOM_FAVORITE packet identifier.
	Header uint16 = 3817
)

// Payload contains the unpacked ROOM_FAVORITE fields.
type Payload struct {
	// RoomID identifies the room to add.
	RoomID int32
}

// Definition describes the ROOM_FAVORITE payload fields.
var Definition = codec.Definition{codec.Named("roomId", codec.Int32Field)}

// Decode unpacks a ROOM_FAVORITE packet payload.
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
