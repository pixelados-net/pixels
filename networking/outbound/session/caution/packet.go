// Package caution contains the MODERATION_CAUTION outbound packet.
package caution

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the MODERATION_CAUTION packet identifier.
	Header uint16 = 1890
)

// Definition describes the MODERATION_CAUTION payload fields.
var Definition = codec.Definition{
	codec.Named("field1", codec.StringField),
	codec.Named("field2", codec.StringField),
}

// Encode creates a MODERATION_CAUTION packet.
func Encode(field1 string, field2 string) (codec.Packet, error) {
	values := make([]codec.Value, 0, 2)
	values = append(values, codec.String(field1))
	values = append(values, codec.String(field2))

	return codec.NewPacket(Header, Definition, values...)
}
