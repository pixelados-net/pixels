// Package favorites contains the USER_FAVORITE_ROOM_COUNT outbound packet.
package favorites

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the USER_FAVORITE_ROOM_COUNT packet identifier.
	Header uint16 = 151
)

// Definition describes the USER_FAVORITE_ROOM_COUNT payload fields.
var Definition = codec.Definition{
	codec.Named("limit", codec.Int32Field),
	codec.Named("roomCount", codec.Int32Field),
}

// RoomDefinition describes one favorite room id entry.
var RoomDefinition = codec.Definition{codec.Named("roomId", codec.Int32Field)}

// Encode creates a USER_FAVORITE_ROOM_COUNT packet.
func Encode(limit int32, roomIDs []int32) (codec.Packet, error) {
	payload, err := codec.AppendPayload(nil, Definition, codec.Int32(limit), codec.Int32(int32(len(roomIDs))))
	if err != nil {
		return codec.Packet{}, err
	}
	for _, roomID := range roomIDs {
		payload, err = codec.AppendPayload(payload, RoomDefinition, codec.Int32(roomID))
		if err != nil {
			return codec.Packet{}, err
		}
	}

	return codec.Packet{Header: Header, Payload: payload}, nil
}
