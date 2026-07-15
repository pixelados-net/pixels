// Package releaseissues contains moderation issue release.
package releaseissues

import "github.com/niflaot/pixels/networking/codec"

// Header identifies RELEASE_ISSUES.
const Header uint16 = 1572

// Decode returns bounded issue ids.
func Decode(packet codec.Packet) ([]int32, error) {
	if packet.Header != Header {
		return nil, codec.ErrUnexpectedHeader
	}
	return decodeIDs(packet.Payload)
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
