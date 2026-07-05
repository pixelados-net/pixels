// Package complete contains the HANDSHAKE_COMPLETE_DIFFIE outbound packet.
package complete

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the HANDSHAKE_COMPLETE_DIFFIE packet identifier.
	Header uint16 = 3885
)

// Definition describes the HANDSHAKE_COMPLETE_DIFFIE payload fields.
var Definition = codec.Definition{
	codec.Named("encryptedPublicKey", codec.StringField),
	codec.Optional(codec.Named("serverClientEncryption", codec.BooleanField)),
}

// Option configures optional HANDSHAKE_COMPLETE_DIFFIE fields.
type Option func(*options)

// options stores optional HANDSHAKE_COMPLETE_DIFFIE fields.
type options struct {
	ServerClientEncryption *bool
}

// WithServerClientEncryption includes the serverClientEncryption protocol field.
func WithServerClientEncryption(value bool) Option {
	return func(options *options) {
		options.ServerClientEncryption = &value
	}
}

// Encode creates a HANDSHAKE_COMPLETE_DIFFIE packet.
func Encode(encryptedPublicKey string, opts ...Option) (codec.Packet, error) {
	payload := options{}
	for _, option := range opts {
		option(&payload)
	}

	values := make([]codec.Value, 0, 2)
	values = append(values, codec.String(encryptedPublicKey))
	if payload.ServerClientEncryption != nil {
		values = append(values, codec.Bool(*payload.ServerClientEncryption))
	}

	return codec.NewPacket(Header, Definition, values...)
}
