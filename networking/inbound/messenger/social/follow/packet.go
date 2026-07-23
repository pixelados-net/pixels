// Package follow contains FOLLOW_FRIEND.
package follow

import "github.com/niflaot/pixels/networking/codec"

// Header identifies FOLLOW_FRIEND.
const Header uint16 = 3997

// Decode unpacks the target player id.
func Decode(packet codec.Packet) (int64, error) {
	if packet.Header != Header {
		return 0, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, codec.Definition{codec.Named("playerId", codec.Int32Field)})
	if err != nil {
		return 0, err
	}
	return int64(values[0].Int32), nil
}
