// Package songinfo decodes the GET_SONG_INFO inbound request.
package songinfo

import (
	"github.com/niflaot/pixels/networking/codec"
)

// Header identifies GET_SONG_INFO.
const Header uint16 = 3082

// MaxSongIDs bounds one client request.
const MaxSongIDs int32 = 100

// Decode validates the requested song identifiers.
func Decode(packet codec.Packet) ([]int32, error) {
	if packet.Header != Header {
		return nil, codec.ErrUnexpectedHeader
	}
	values, rest, err := codec.DecodePacket(packet, codec.Definition{codec.Int32Field})
	if err != nil || values[0].Int32 < 0 || values[0].Int32 > MaxSongIDs {
		return nil, codec.ErrInvalidField
	}
	ids := make([]int32, values[0].Int32)
	for index := range ids {
		values, rest, err = codec.DecodePayload(nil, codec.Definition{codec.Int32Field}, rest)
		if err != nil {
			return nil, err
		}
		ids[index] = values[0].Int32
	}
	if len(rest) != 0 {
		return nil, codec.ErrUnexpectedPayload
	}
	return ids, nil
}
