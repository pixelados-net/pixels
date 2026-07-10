// Package mute contains the ROOM_MUTE_USER inbound packet.
package mute

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header identifies ROOM_MUTE_USER.
	Header uint16 = 3485
)

// Payload contains unpacked room mute fields.
type Payload struct {
	// PlayerID identifies the target player.
	PlayerID int32
	// RoomID identifies the room.
	RoomID int32
	// Minutes stores the requested duration, or zero to unmute.
	Minutes int32
}

// Definition describes ROOM_MUTE_USER fields.
var Definition = codec.Definition{
	codec.Named("playerId", codec.Int32Field),
	codec.Named("roomId", codec.Int32Field),
	codec.Named("minutes", codec.Int32Field),
}

// Decode unpacks a ROOM_MUTE_USER packet.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return Payload{}, err
	}

	return Payload{PlayerID: values[0].Int32, RoomID: values[1].Int32, Minutes: values[2].Int32}, nil
}
