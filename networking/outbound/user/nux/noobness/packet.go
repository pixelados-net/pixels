// Package noobness contains the NOOBNESS_LEVEL outbound packet.
package noobness

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header identifies NOOBNESS_LEVEL.
	Header uint16 = 3738
	// OldIdentity disables the retired NUX journey.
	OldIdentity int32 = 0
)

// Definition describes NOOBNESS_LEVEL fields.
var Definition = codec.Definition{codec.Named("level", codec.Int32Field)}

// Encode creates a NOOBNESS_LEVEL packet.
func Encode(level int32) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.Int32(level))
}
