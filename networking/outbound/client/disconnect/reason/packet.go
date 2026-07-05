// Package reason contains the DISCONNECT_REASON outbound packet.
package reason

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the DISCONNECT_REASON packet identifier.
	Header uint16 = 4000
)

// Definition describes the DISCONNECT_REASON payload fields.
var Definition = codec.Definition{
	codec.Optional(codec.Named("reason", codec.Int32Field)),
}

// Option configures optional DISCONNECT_REASON fields.
type Option func(*options)

// options stores optional DISCONNECT_REASON fields.
type options struct {
	Reason *int32
}

// WithReason includes the reason protocol field.
func WithReason(value int32) Option {
	return func(options *options) {
		options.Reason = &value
	}
}

// Encode creates a DISCONNECT_REASON packet.
func Encode(opts ...Option) (codec.Packet, error) {
	payload := options{}
	for _, option := range opts {
		option(&payload)
	}

	values := make([]codec.Value, 0, 1)
	if payload.Reason != nil {
		values = append(values, codec.Int32(*payload.Reason))
	}

	return codec.NewPacket(Header, Definition, values...)
}
