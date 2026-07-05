// Package nosuchflat contains the NO_SUCH_FLAT outbound packet.
package nosuchflat

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the NO_SUCH_FLAT packet identifier.
	Header uint16 = 84
)

// Definition describes the NO_SUCH_FLAT payload fields.
var Definition = codec.Definition{codec.Named("field1", codec.Int32Field)}

// Encode creates a NO_SUCH_FLAT packet.
func Encode(value int32) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.Int32(value))
}
