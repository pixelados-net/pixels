// Package maintenance contains the HOTEL_MAINTENANCE outbound packet.
package maintenance

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the HOTEL_MAINTENANCE packet identifier.
	Header uint16 = 1350
)

// Definition describes the HOTEL_MAINTENANCE payload fields.
var Definition = codec.Definition{
	codec.Named("isInMaintenance", codec.BooleanField),
	codec.Named("minutesUntilMaintenance", codec.Int32Field),
	codec.Optional(codec.Named("duration", codec.Int32Field)),
}

// Option configures optional HOTEL_MAINTENANCE fields.
type Option func(*options)

// options stores optional HOTEL_MAINTENANCE fields.
type options struct {
	Duration *int32
}

// WithDuration includes the duration protocol field.
func WithDuration(value int32) Option {
	return func(options *options) {
		options.Duration = &value
	}
}

// Encode creates a HOTEL_MAINTENANCE packet.
func Encode(isInMaintenance bool, minutesUntilMaintenance int32, opts ...Option) (codec.Packet, error) {
	payload := options{}
	for _, option := range opts {
		option(&payload)
	}

	values := make([]codec.Value, 0, 3)
	values = append(values, codec.Bool(isInMaintenance))
	values = append(values, codec.Int32(minutesUntilMaintenance))
	if payload.Duration != nil {
		values = append(values, codec.Int32(*payload.Duration))
	}

	return codec.NewPacket(Header, Definition, values...)
}
