// Package network decodes the retired ROOM_DIRECTORY_ROOM_NETWORK_OPEN_CONNECTION request.
package network

import (
	"github.com/niflaot/pixels/networking/codec"
	"github.com/niflaot/pixels/networking/inbound"
)

// Header identifies ROOM_DIRECTORY_ROOM_NETWORK_OPEN_CONNECTION.
const Header uint16 = 3736

// Payload contains the abandoned room-network edge.
type Payload struct {
	// RoomID identifies the source room.
	RoomID int32
	// TargetRoomID identifies the target room.
	TargetRoomID int32
}

// Definition describes the two room ids.
var Definition = codec.Definition{codec.Named("roomId", codec.Int32Field), codec.Named("targetRoomId", codec.Int32Field)}

// Decode returns the compatibility-only room ids.
func Decode(packet codec.Packet) (Payload, error) {
	if err := inbound.ValidateHeader(packet, Header); err != nil {
		return Payload{}, err
	}
	v, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return Payload{}, err
	}
	return Payload{RoomID: v[0].Int32, TargetRoomID: v[1].Int32}, nil
}
