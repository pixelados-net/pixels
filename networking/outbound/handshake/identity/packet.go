// Package identity contains the HANDSHAKE_IDENTITY_ACCOUNT outbound packet.
package identity

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the HANDSHAKE_IDENTITY_ACCOUNT packet identifier.
	Header uint16 = 3523
)

// Definition describes the HANDSHAKE_IDENTITY_ACCOUNT payload fields.
var Definition = codec.Definition{
	codec.Named("count", codec.Int32Field),
}

// Encode creates a HANDSHAKE_IDENTITY_ACCOUNT packet.
func Encode(count int32) (codec.Packet, error) {
	values := make([]codec.Value, 0, 1)
	values = append(values, codec.Int32(count))

	return codec.NewPacket(Header, Definition, values...)
}
