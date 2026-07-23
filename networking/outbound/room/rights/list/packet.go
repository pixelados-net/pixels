// Package list contains the ROOM_RIGHTS_LIST outbound packet.
package list

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header identifies ROOM_RIGHTS_LIST.
	Header uint16 = 1284
)

// Right contains one protocol room controller.
type Right struct {
	// PlayerID identifies the controller.
	PlayerID int32
	// Username stores the controller username.
	Username string
}

// Definition describes list metadata.
var Definition = codec.Definition{codec.Named("roomId", codec.Int32Field), codec.Named("count", codec.Int32Field)}

// RightDefinition describes one rights holder.
var RightDefinition = codec.Definition{codec.Named("playerId", codec.Int32Field), codec.Named("username", codec.StringField)}

// Encode creates a ROOM_RIGHTS_LIST packet.
func Encode(roomID int32, rights []Right) (codec.Packet, error) {
	payload, err := codec.AppendPayload(nil, Definition, codec.Int32(roomID), codec.Int32(int32(len(rights))))
	if err != nil {
		return codec.Packet{}, err
	}
	for _, right := range rights {
		payload, err = codec.AppendPayload(payload, RightDefinition, codec.Int32(right.PlayerID), codec.String(right.Username))
		if err != nil {
			return codec.Packet{}, err
		}
	}

	return codec.Packet{Header: Header, Payload: payload}, nil
}
