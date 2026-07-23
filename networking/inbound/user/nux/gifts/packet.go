// Package gifts contains the retired NEW_USER_EXPERIENCE_GET_GIFTS packet.
package gifts

import (
	"github.com/niflaot/pixels/networking/codec"
)

const (
	// Header identifies NEW_USER_EXPERIENCE_GET_GIFTS.
	Header uint16 = 1822
	// MaxValues bounds the retired selection payload before allocation.
	MaxValues int32 = 90
)

// CountDefinition describes the NUX integer count prefix.
var CountDefinition = codec.Definition{codec.Named("valueCount", codec.Int32Field)}

// ValueDefinition describes one integer in the retired NUX selection triples.
var ValueDefinition = codec.Definition{codec.Named("value", codec.Int32Field)}

// Payload contains decoded NUX gift integer triples.
type Payload struct {
	// Values stores day, step, and gift integers in wire order.
	Values []int32
}

// Decode decodes and bounds a retired NUX gift packet.
//
// Deprecated: the legacy NUX journey is intentionally retired.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, rest, err := codec.DecodePacket(packet, CountDefinition)
	if err != nil {
		return Payload{}, err
	}
	count := values[0].Int32
	if count < 0 || count > MaxValues || count%3 != 0 {
		return Payload{}, codec.ErrInvalidField
	}
	result := Payload{Values: make([]int32, 0, count)}
	for range count {
		decoded, remaining, decodeErr := codec.DecodePayload(nil, ValueDefinition, rest)
		if decodeErr != nil {
			return Payload{}, decodeErr
		}
		result.Values = append(result.Values, decoded[0].Int32)
		rest = remaining
	}
	if len(rest) != 0 {
		return Payload{}, codec.ErrUnexpectedPayload
	}
	return result, nil
}
