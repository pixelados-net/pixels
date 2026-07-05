// Package status contains the AVAILABILITY_STATUS outbound packet.
package status

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the AVAILABILITY_STATUS packet identifier.
	Header uint16 = 2033
)

// Definition describes the AVAILABILITY_STATUS payload fields.
var Definition = codec.Definition{
	codec.Named("isOpen", codec.BooleanField),
	codec.Named("onShutdown", codec.BooleanField),
	codec.Optional(codec.Named("isAuthentic", codec.BooleanField)),
}

// Option configures optional AVAILABILITY_STATUS fields.
type Option func(*options)

// options stores optional AVAILABILITY_STATUS fields.
type options struct {
	IsAuthentic *bool
}

// WithIsAuthentic includes the isAuthentic protocol field.
func WithIsAuthentic(value bool) Option {
	return func(options *options) {
		options.IsAuthentic = &value
	}
}

// Encode creates a AVAILABILITY_STATUS packet.
func Encode(isOpen bool, onShutdown bool, opts ...Option) (codec.Packet, error) {
	payload := options{}
	for _, option := range opts {
		option(&payload)
	}

	values := make([]codec.Value, 0, 3)
	values = append(values, codec.Bool(isOpen))
	values = append(values, codec.Bool(onShutdown))
	if payload.IsAuthentic != nil {
		values = append(values, codec.Bool(*payload.IsAuthentic))
	}

	return codec.NewPacket(Header, Definition, values...)
}
