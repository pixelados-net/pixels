// Package decline contains DECLINE_FRIEND.
package decline

import "github.com/niflaot/pixels/networking/codec"

// Header identifies DECLINE_FRIEND.
const Header uint16 = 2890

// Payload contains decline mode and requester ids.
type Payload struct {
	// All reports whether every incoming request should be declined.
	All bool
	// PlayerIDs identifies selected requesters.
	PlayerIDs []int64
}

// Decode unpacks DECLINE_FRIEND.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, rest, err := codec.DecodePayload(nil, codec.Definition{codec.BooleanField, codec.Int32Field}, packet.Payload)
	if err != nil {
		return Payload{}, err
	}
	count := int(values[1].Int32)
	if count < 0 || count > 100 {
		return Payload{}, codec.ErrInvalidField
	}
	result := Payload{All: values[0].Boolean, PlayerIDs: make([]int64, 0, count)}
	for range count {
		values, rest, err = codec.DecodePayload(values[:0], codec.Definition{codec.Int32Field}, rest)
		if err != nil {
			return Payload{}, err
		}
		result.PlayerIDs = append(result.PlayerIDs, int64(values[0].Int32))
	}
	if len(rest) != 0 {
		return Payload{}, codec.ErrUnexpectedPayload
	}
	return result, nil
}
