// Package bubblealert contains the BUBBLE_ALERT outbound packet.
package bubblealert

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the BUBBLE_ALERT packet identifier.
	Header uint16 = 1992
)

// Pair stores one bubble alert option.
type Pair struct {
	// Key stores the option name.
	Key string
	// Value stores the option value.
	Value string
}

// Option configures a bubble alert packet.
type Option func(*params)

// params stores bubble alert options.
type params struct {
	// pairs stores option pairs after message.
	pairs []Pair
}

// Definition describes the base BUBBLE_ALERT payload fields.
var Definition = codec.Definition{
	codec.Named("errorKey", codec.StringField),
	codec.Named("keyCount", codec.Int32Field),
	codec.Named("key", codec.StringField),
	codec.Named("value", codec.StringField),
}

// WithParam appends a custom option pair.
func WithParam(key string, value string) Option {
	return func(params *params) {
		params.pairs = append(params.pairs, Pair{Key: key, Value: value})
	}
}

// WithDisplayBubble requests Nitro bubble presentation.
func WithDisplayBubble() Option {
	return WithParam("display", "BUBBLE")
}

// Encode creates a BUBBLE_ALERT packet.
func Encode(errorKey string, message string, options ...Option) (codec.Packet, error) {
	parameters := params{}
	for _, option := range options {
		option(&parameters)
	}

	pairs := append([]Pair{{Key: "message", Value: message}}, parameters.pairs...)
	values := make([]codec.Value, 0, 2+len(pairs)*2)
	values = append(values, codec.String(errorKey), codec.Int32(int32(len(pairs))))
	for _, pair := range pairs {
		values = append(values, codec.String(pair.Key), codec.String(pair.Value))
	}

	return codec.NewPacket(Header, definition(len(pairs)), values...)
}

// definition returns a BUBBLE_ALERT definition for pair count.
func definition(count int) codec.Definition {
	fields := make(codec.Definition, 0, 2+count*2)
	fields = append(fields,
		codec.Named("errorKey", codec.StringField),
		codec.Named("keyCount", codec.Int32Field),
	)
	for index := 0; index < count; index++ {
		fields = append(fields,
			codec.Named("key", codec.StringField),
			codec.Named("value", codec.StringField),
		)
	}

	return fields
}
