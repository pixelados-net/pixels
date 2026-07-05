// Package cancreate contains the CAN_CREATE_ROOM outbound packet.
package cancreate

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the CAN_CREATE_ROOM packet identifier.
	Header uint16 = 378
)

// Definition describes the CAN_CREATE_ROOM payload fields.
var Definition = codec.Definition{
	codec.Named("resultCode", codec.Int32Field),
	codec.Named("roomLimit", codec.Int32Field),
}

// Encode creates a CAN_CREATE_ROOM packet.
func Encode(resultCode int32, roomLimit int32) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.Int32(resultCode), codec.Int32(roomLimit))
}
