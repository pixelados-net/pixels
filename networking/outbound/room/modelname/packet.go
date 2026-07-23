// Package modelname contains the ROOM_MODEL_NAME outbound packet.
package modelname

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the ROOM_MODEL_NAME packet identifier.
	Header uint16 = 2031
)

// Definition describes the ROOM_MODEL_NAME payload fields.
var Definition = codec.Definition{
	codec.Named("modelName", codec.StringField),
	codec.Named("roomId", codec.Int32Field),
}

// Encode creates a ROOM_MODEL_NAME packet.
func Encode(modelName string, roomID int32) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.String(modelName), codec.Int32(roomID))
}
