// Package entryerror contains the ROOM_ENTER_ERROR outbound packet.
package entryerror

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the ROOM_ENTER_ERROR packet identifier.
	Header uint16 = 899
)

// Option configures optional ROOM_ENTER_ERROR fields.
type Option func(*options)

// options stores optional ROOM_ENTER_ERROR values.
type options struct {
	// parameter stores an optional queue error parameter.
	parameter string
}

// Definition describes the ROOM_ENTER_ERROR payload fields.
var Definition = codec.Definition{
	codec.Named("errorCode", codec.Int32Field),
	codec.Named("parameter", codec.StringField),
}

// Encode creates a ROOM_ENTER_ERROR packet.
func Encode(errorCode int32, optionFunctions ...Option) (codec.Packet, error) {
	configured := options{}
	for _, option := range optionFunctions {
		option(&configured)
	}

	return codec.NewPacket(Header, Definition, codec.Int32(errorCode), codec.String(configured.parameter))
}

// WithParameter configures a queue error parameter.
func WithParameter(parameter string) Option {
	return func(configured *options) {
		configured.parameter = parameter
	}
}
