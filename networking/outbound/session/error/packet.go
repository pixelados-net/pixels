// Package error contains the GENERIC_ERROR outbound packet.
package error

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the GENERIC_ERROR packet identifier.
	Header uint16 = 1600
)

// Definition describes the GENERIC_ERROR payload fields.
var Definition = codec.Definition{
	codec.Named("field1", codec.Int32Field),
}

// Encode creates a GENERIC_ERROR packet.
func Encode(field1 int32) (codec.Packet, error) {
	values := make([]codec.Value, 0, 1)
	values = append(values, codec.Int32(field1))

	return codec.NewPacket(Header, Definition, values...)
}
