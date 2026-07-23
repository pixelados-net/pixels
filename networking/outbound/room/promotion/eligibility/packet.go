// Package eligibility encodes CAN_CREATE_ROOM_EVENT responses.
package eligibility

import "github.com/niflaot/pixels/networking/codec"

// Header identifies CAN_CREATE_ROOM_EVENT.
const Header uint16 = 2599

// Definition describes eligibility and its error code.
var Definition = codec.Definition{codec.Named("canCreate", codec.BooleanField), codec.Named("errorCode", codec.Int32Field)}

// Encode creates one room-event eligibility response.
func Encode(canCreate bool, errorCode int32) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.Bool(canCreate), codec.Int32(errorCode))
}
