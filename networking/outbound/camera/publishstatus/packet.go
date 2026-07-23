// Package publishstatus contains CAMERA_PUBLISH_STATUS.
package publishstatus

import "github.com/niflaot/pixels/networking/codec"

// Header identifies CAMERA_PUBLISH_STATUS.
const Header uint16 = 2057

// Option configures successful optional fields.
type Option func(*result)

// result stores optional publish fields.
type result struct {
	// url stores the durable publication URL.
	url string
}

// WithURL supplies a successful durable publication URL.
func WithURL(url string) Option { return func(value *result) { value.url = url } }

// Encode creates a publish result with a conditional URL.
func Encode(ok bool, cooldownSeconds int32, options ...Option) (codec.Packet, error) {
	if cooldownSeconds < 0 {
		return codec.Packet{}, codec.ErrInvalidField
	}
	payload, err := codec.AppendPayload(nil, codec.Definition{codec.BooleanField, codec.Int32Field}, codec.Bool(ok), codec.Int32(cooldownSeconds))
	if err != nil || !ok {
		return codec.Packet{Header: Header, Payload: payload}, err
	}
	value := result{}
	for _, option := range options {
		option(&value)
	}
	if value.url == "" {
		return codec.Packet{Header: Header, Payload: payload}, nil
	}
	payload, err = codec.AppendPayload(payload, codec.Definition{codec.StringField}, codec.String(value.url))
	return codec.Packet{Header: Header, Payload: payload}, err
}
