// Package ban contains the ROOM_BAN_GIVE inbound packet.
package ban

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header identifies ROOM_BAN_GIVE.
	Header uint16 = 1477
)

// Payload contains unpacked room ban fields.
type Payload struct {
	// PlayerID identifies the target player.
	PlayerID int32
	// RoomID identifies the room.
	RoomID int32
	// Duration stores the Nitro duration name.
	Duration string
}

// Definition describes ROOM_BAN_GIVE fields.
var Definition = codec.Definition{
	codec.Named("playerId", codec.Int32Field),
	codec.Named("roomId", codec.Int32Field),
	codec.Named("duration", codec.StringField),
}

// Decode unpacks a ROOM_BAN_GIVE packet.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return Payload{}, err
	}

	return Payload{PlayerID: values[0].Int32, RoomID: values[1].Int32, Duration: values[2].String}, nil
}
