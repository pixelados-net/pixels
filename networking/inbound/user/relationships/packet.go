// Package relationships decodes MESSENGER_RELATIONSHIPS requests.
package relationships

import "github.com/niflaot/pixels/networking/codec"

// Header identifies MESSENGER_RELATIONSHIPS.
const Header uint16 = 2138

// Decode unpacks the target player id.
func Decode(packet codec.Packet) (int64, error) {
	if packet.Header != Header {
		return 0, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, codec.Definition{codec.Int32Field})
	if err != nil {
		return 0, err
	}
	return int64(values[0].Int32), nil
}
