// Package purchase encodes ROOM_AD_PURCHASE information responses.
package purchase

import "github.com/niflaot/pixels/networking/codec"

// Header identifies ROOM_AD_PURCHASE.
const Header uint16 = 2468

// Room contains one player-owned room eligible for promotion.
type Room struct {
	// ID identifies the room.
	ID int32
	// Name stores the visible room name.
	Name string
	// Promoted reports whether the room already has an active promotion.
	Promoted bool
}

// Definition describes the VIP flag and room count prefix.
var Definition = codec.Definition{codec.Named("vip", codec.BooleanField), codec.Named("roomCount", codec.Int32Field)}

// RoomDefinition describes one eligible room.
var RoomDefinition = codec.Definition{codec.Named("roomId", codec.Int32Field), codec.Named("roomName", codec.StringField), codec.Named("promoted", codec.BooleanField)}

// Encode creates one room-ad purchase information response.
func Encode(vip bool, rooms []Room) (codec.Packet, error) {
	payload, err := codec.AppendPayload(nil, Definition, codec.Bool(vip), codec.Int32(int32(len(rooms))))
	for index := 0; err == nil && index < len(rooms); index++ {
		room := rooms[index]
		payload, err = codec.AppendPayload(payload, RoomDefinition, codec.Int32(room.ID), codec.String(room.Name), codec.Bool(room.Promoted))
	}
	return codec.Packet{Header: Header, Payload: payload}, err
}
