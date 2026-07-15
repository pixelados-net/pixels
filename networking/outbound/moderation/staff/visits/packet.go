// Package visits contains MODTOOL_VISITED_ROOMS_USER projection.
package visits

import "github.com/niflaot/pixels/networking/codec"

// Header identifies MODTOOL_VISITED_ROOMS_USER.
const Header uint16 = 1752

// Visit stores one room entry.
type Visit struct {
	RoomID   int32
	RoomName string
	Hour     int32
	Minute   int32
}

// Encode creates a visited-room list.
func Encode(playerID int32, username string, visits []Visit) (codec.Packet, error) {
	payload, err := codec.AppendPayload(nil, codec.Definition{codec.Int32Field, codec.StringField, codec.Int32Field}, codec.Int32(playerID), codec.String(username), codec.Int32(int32(len(visits))))
	if err != nil {
		return codec.Packet{}, err
	}
	for _, visit := range visits {
		payload, err = codec.AppendPayload(payload, codec.Definition{codec.Int32Field, codec.StringField, codec.Int32Field, codec.Int32Field}, codec.Int32(visit.RoomID), codec.String(visit.RoomName), codec.Int32(visit.Hour), codec.Int32(visit.Minute))
		if err != nil {
			return codec.Packet{}, err
		}
	}
	return codec.Packet{Header: Header, Payload: payload}, nil
}
