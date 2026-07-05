// Package entryerror contains the ROOM_ENTER_ERROR outbound packet.
package entryerror

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the ROOM_ENTER_ERROR packet identifier.
	Header uint16 = 899
)

// Definition describes the ROOM_ENTER_ERROR payload fields.
var Definition = codec.Definition{codec.Named("errorCode", codec.Int32Field)}

// Encode creates a ROOM_ENTER_ERROR packet.
func Encode(errorCode int32) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.Int32(errorCode))
}
