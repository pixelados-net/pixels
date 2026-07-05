// Package machine contains the SECURITY_MACHINE outbound packet.
package machine

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the SECURITY_MACHINE packet identifier.
	Header uint16 = 1488
)

// Definition describes the SECURITY_MACHINE payload fields.
var Definition = codec.Definition{
	codec.Named("machineId", codec.StringField),
}

// Encode creates a SECURITY_MACHINE packet.
func Encode(machineID string) (codec.Packet, error) {
	values := make([]codec.Value, 0, 1)
	values = append(values, codec.String(machineID))

	return codec.NewPacket(Header, Definition, values...)
}
