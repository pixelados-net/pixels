// Package unban contains the ROOM_BAN_REMOVE inbound packet.
package unban

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header identifies ROOM_BAN_REMOVE.
	Header uint16 = 992
)

// Payload contains unpacked room unban fields.
type Payload struct {
	// PlayerID identifies the target player.
	PlayerID int32
	// RoomID identifies the room.
	RoomID int32
}

// Definition describes ROOM_BAN_REMOVE fields.
var Definition = codec.Definition{codec.Named("playerId", codec.Int32Field), codec.Named("roomId", codec.Int32Field)}

// Decode unpacks a ROOM_BAN_REMOVE packet.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return Payload{}, err
	}

	return Payload{PlayerID: values[0].Int32, RoomID: values[1].Int32}, nil
}
