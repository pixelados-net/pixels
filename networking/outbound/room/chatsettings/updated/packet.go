// Package updated contains the ROOM_SETTINGS_CHAT outbound packet.
package updated

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header identifies ROOM_SETTINGS_CHAT.
	Header uint16 = 1191
)

// Definition describes ROOM_SETTINGS_CHAT fields.
var Definition = codec.Definition{codec.Named("mode", codec.Int32Field), codec.Named("weight", codec.Int32Field), codec.Named("speed", codec.Int32Field), codec.Named("distance", codec.Int32Field), codec.Named("protection", codec.Int32Field)}

// Encode creates a ROOM_SETTINGS_CHAT packet.
func Encode(mode int32, weight int32, speed int32, distance int32, protection int32) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.Int32(mode), codec.Int32(weight), codec.Int32(speed), codec.Int32(distance), codec.Int32(protection))
}
