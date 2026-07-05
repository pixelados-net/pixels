// Package feed contains the INFO_FEED_ENABLE outbound packet.
package feed

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the INFO_FEED_ENABLE packet identifier.
	Header uint16 = 3284
)

// Definition describes the INFO_FEED_ENABLE payload fields.
var Definition = codec.Definition{
	codec.Named("field1", codec.BooleanField),
}

// Encode creates a INFO_FEED_ENABLE packet.
func Encode(field1 bool) (codec.Packet, error) {
	values := make([]codec.Value, 0, 1)
	values = append(values, codec.Bool(field1))

	return codec.NewPacket(Header, Definition, values...)
}
