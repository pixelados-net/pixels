// Package closeissues contains moderation issue resolution.
package closeissues

import "github.com/niflaot/pixels/networking/codec"

// Header identifies CLOSE_ISSUES.
const Header uint16 = 2067

// Payload contains resolution and issue ids.
type Payload struct {
	Resolution int32
	IssueIDs   []int32
}

// Decode reads resolution followed by a bounded list.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, rest, err := codec.DecodePayload(nil, codec.Definition{codec.Int32Field}, packet.Payload)
	if err != nil {
		return Payload{}, err
	}
	ids, err := decodeIDs(rest)
	return Payload{Resolution: values[0].Int32, IssueIDs: ids}, err
}

// decodeIDs reads an int32-counted list.
func decodeIDs(payload []byte) ([]int32, error) {
	values, rest, err := codec.DecodePayload(nil, codec.Definition{codec.Int32Field}, payload)
	if err != nil || values[0].Int32 < 0 || values[0].Int32 > 100 {
		return nil, codec.ErrInvalidField
	}
	ids := make([]int32, values[0].Int32)
	for i := range ids {
		values, rest, err = codec.DecodePayload(values[:0], codec.Definition{codec.Int32Field}, rest)
		if err != nil {
			return nil, err
		}
		ids[i] = values[0].Int32
	}
	if len(rest) != 0 {
		return nil, codec.ErrUnexpectedPayload
	}
	return ids, nil
}
