// Package score contains the ROOM_SCORE outbound packet.
package score

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header identifies ROOM_SCORE.
	Header uint16 = 482
)

// Definition describes ROOM_SCORE fields.
var Definition = codec.Definition{
	codec.Named("score", codec.Int32Field),
	codec.Named("canLike", codec.BooleanField),
}

// Encode creates a ROOM_SCORE packet.
func Encode(value int32, canLike bool) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.Int32(value), codec.Bool(canLike))
}
