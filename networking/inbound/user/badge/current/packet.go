// Package current decodes current badge requests for room users.
package current

import "github.com/niflaot/pixels/networking/codec"

// Header identifies USER_BADGES_CURRENT requests.
const Header uint16 = 2091

// Definition describes the requested player identifier.
var Definition = codec.Definition{codec.Named("playerId", codec.Int32Field)}

// Decode returns the requested player identifier.
func Decode(packet codec.Packet) (int64, error) {
	if packet.Header != Header {
		return 0, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return 0, err
	}
	return int64(values[0].Int32), nil
}
