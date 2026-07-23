// Package remove contains REMOVE_FRIEND.
package remove

import "github.com/niflaot/pixels/networking/codec"

// Header identifies REMOVE_FRIEND.
const Header uint16 = 1689

// Decode unpacks a bounded friend id batch.
func Decode(packet codec.Packet) ([]int64, error) {
	if packet.Header != Header {
		return nil, codec.ErrUnexpectedHeader
	}
	values, rest, err := codec.DecodePayload(nil, codec.Definition{codec.Int32Field}, packet.Payload)
	if err != nil {
		return nil, err
	}
	count := int(values[0].Int32)
	if count < 0 || count > 100 {
		return nil, codec.ErrInvalidField
	}
	ids := make([]int64, 0, count)
	for range count {
		values, rest, err = codec.DecodePayload(values[:0], codec.Definition{codec.Int32Field}, rest)
		if err != nil {
			return nil, err
		}
		ids = append(ids, int64(values[0].Int32))
	}
	if len(rest) != 0 {
		return nil, codec.ErrUnexpectedPayload
	}
	return ids, nil
}
