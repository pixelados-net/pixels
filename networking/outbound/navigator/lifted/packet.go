// Package lifted contains the NAVIGATOR_LIFTED outbound packet.
package lifted

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the NAVIGATOR_LIFTED packet identifier.
	Header uint16 = 3104
)

// Room contains one lifted room entry.
type Room struct {
	// RoomID identifies the lifted room.
	RoomID int32
	// AreaID identifies the navigator area.
	AreaID int32
	// Image stores the lifted image reference.
	Image string
	// Caption stores the lifted room caption.
	Caption string
}

// Definition describes the NAVIGATOR_LIFTED payload fields.
var Definition = codec.Definition{codec.Named("roomCount", codec.Int32Field)}

// RoomDefinition describes one lifted room entry.
var RoomDefinition = codec.Definition{
	codec.Named("roomId", codec.Int32Field),
	codec.Named("areaId", codec.Int32Field),
	codec.Named("image", codec.StringField),
	codec.Named("caption", codec.StringField),
}

// Encode creates a NAVIGATOR_LIFTED packet.
func Encode(rooms []Room) (codec.Packet, error) {
	payload, err := codec.AppendPayload(nil, Definition, codec.Int32(int32(len(rooms))))
	if err != nil {
		return codec.Packet{}, err
	}
	for _, room := range rooms {
		payload, err = codec.AppendPayload(payload, RoomDefinition,
			codec.Int32(room.RoomID),
			codec.Int32(room.AreaID),
			codec.String(room.Image),
			codec.String(room.Caption),
		)
		if err != nil {
			return codec.Packet{}, err
		}
	}

	return codec.Packet{Header: Header, Payload: payload}, nil
}
