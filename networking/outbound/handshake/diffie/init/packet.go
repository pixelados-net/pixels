// Package init contains the HANDSHAKE_INIT_DIFFIE outbound packet.
package init

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the HANDSHAKE_INIT_DIFFIE packet identifier.
	Header uint16 = 1347
)

// Definition describes the HANDSHAKE_INIT_DIFFIE payload fields.
var Definition = codec.Definition{
	codec.Named("encryptedPrime", codec.StringField),
	codec.Named("encryptedGenerator", codec.StringField),
}

// Encode creates a HANDSHAKE_INIT_DIFFIE packet.
func Encode(encryptedPrime string, encryptedGenerator string) (codec.Packet, error) {
	values := make([]codec.Value, 0, 2)
	values = append(values, codec.String(encryptedPrime))
	values = append(values, codec.String(encryptedGenerator))

	return codec.NewPacket(Header, Definition, values...)
}
