// Package dance encodes UNIT_DANCE broadcasts.
package dance

import "github.com/niflaot/pixels/networking/codec"

// Header identifies UNIT_DANCE.
const Header uint16 = 2233

// Definition describes UNIT_DANCE fields.
var Definition = codec.Definition{codec.Named("roomIndex", codec.Int32Field), codec.Named("danceId", codec.Int32Field)}

// Encode creates a UNIT_DANCE packet.
func Encode(roomIndex int64, danceID int32) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.Int32(int32(roomIndex)), codec.Int32(danceID))
}
