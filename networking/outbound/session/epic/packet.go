// Package epic contains the EPIC_POPUP outbound packet.
package epic

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the EPIC_POPUP packet identifier.
	Header uint16 = 3945
)

// Definition describes the EPIC_POPUP payload fields.
var Definition = codec.Definition{
	codec.Named("field1", codec.StringField),
}

// Encode creates a EPIC_POPUP packet.
func Encode(field1 string) (codec.Packet, error) {
	values := make([]codec.Value, 0, 1)
	values = append(values, codec.String(field1))

	return codec.NewPacket(Header, Definition, values...)
}
