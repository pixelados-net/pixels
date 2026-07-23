// Package readmarker contains UPDATE_FORUM_READ_MARKER.
package readmarker

import (
	grouprecord "github.com/niflaot/pixels/internal/realm/group/record"
	"github.com/niflaot/pixels/networking/codec"
)

// Header identifies UPDATE_FORUM_READ_MARKER.
const Header uint16 = 1855

// Decode unpacks a bounded collection of marker triples.
func Decode(packet codec.Packet) ([]grouprecord.ReadMarker, error) {
	if packet.Header != Header {
		return nil, codec.ErrUnexpectedHeader
	}
	values, rest, err := codec.DecodePayload(nil, codec.Definition{codec.Int32Field}, packet.Payload)
	if err != nil {
		return nil, err
	}
	count := int(values[0].Int32)
	if count < 0 || count > 50 {
		return nil, codec.ErrInvalidField
	}
	markers := make([]grouprecord.ReadMarker, 0, count)
	for range count {
		values, rest, err = codec.DecodePayload(values[:0], codec.Definition{codec.Int32Field, codec.Int32Field, codec.Int32Field}, rest)
		if err != nil {
			return nil, err
		}
		markers = append(markers, grouprecord.ReadMarker{GroupID: int64(values[0].Int32), LastMessageID: int64(values[1].Int32), Flag: values[2].Int32})
	}
	if len(rest) != 0 {
		return nil, codec.ErrUnexpectedPayload
	}
	return markers, nil
}
