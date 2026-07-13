// Package additem contains the TRADE_ADD_ITEM inbound packet.
package additem

import "github.com/niflaot/pixels/networking/codec"

// Header identifies TRADE_ADD_ITEM.
const Header uint16 = 3107

// Decode reads one item id.
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
