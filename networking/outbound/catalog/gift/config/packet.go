// Package config contains the GIFT_WRAPPER_CONFIG outbound packet.
package config

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header identifies GIFT_WRAPPER_CONFIG.
	Header uint16 = 2234
)

// Options contains available gift wrapping styles.
type Options struct {
	// Price stores the credits surcharge for special wrapping.
	Price int32
	// Wrappers stores available wrapping furniture sprite identifiers.
	Wrappers []int32
	// Boxes stores available box color identifiers.
	Boxes []int32
	// Ribbons stores available ribbon color identifiers.
	Ribbons []int32
	// DefaultGifts stores available free gift furniture sprite identifiers.
	DefaultGifts []int32
}

// Encode creates a GIFT_WRAPPER_CONFIG packet.
func Encode(options Options) (codec.Packet, error) {
	payload, err := codec.AppendPayload(nil, codec.Definition{codec.BooleanField, codec.Int32Field, codec.Int32Field},
		codec.Bool(len(options.Wrappers) != 0 || len(options.DefaultGifts) != 0), codec.Int32(options.Price), codec.Int32(int32(len(options.Wrappers))))
	if err != nil {
		return codec.Packet{}, err
	}
	payload, err = appendIDs(payload, options.Wrappers)
	if err != nil {
		return codec.Packet{}, err
	}
	payload, err = codec.AppendPayload(payload, codec.Definition{codec.Int32Field}, codec.Int32(int32(len(options.Boxes))))
	if err != nil {
		return codec.Packet{}, err
	}
	payload, err = appendIDs(payload, options.Boxes)
	if err != nil {
		return codec.Packet{}, err
	}
	payload, err = codec.AppendPayload(payload, codec.Definition{codec.Int32Field}, codec.Int32(int32(len(options.Ribbons))))
	if err != nil {
		return codec.Packet{}, err
	}
	payload, err = appendIDs(payload, options.Ribbons)
	if err != nil {
		return codec.Packet{}, err
	}
	payload, err = codec.AppendPayload(payload, codec.Definition{codec.Int32Field}, codec.Int32(int32(len(options.DefaultGifts))))
	if err != nil {
		return codec.Packet{}, err
	}
	payload, err = appendIDs(payload, options.DefaultGifts)

	return codec.Packet{Header: Header, Payload: payload}, err
}

// appendIDs appends one list of wrapping identifiers.
func appendIDs(payload []byte, values []int32) ([]byte, error) {
	var err error
	for _, value := range values {
		payload, err = codec.AppendPayload(payload, codec.Definition{codec.Int32Field}, codec.Int32(value))
		if err != nil {
			return payload, err
		}
	}

	return payload, nil
}
