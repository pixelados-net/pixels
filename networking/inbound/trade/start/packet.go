// Package start contains the TRADE_START inbound packet.
package start

import "github.com/niflaot/pixels/networking/codec"

// Header identifies TRADE_START.
const Header uint16 = 1481

// Decode reads the target room unit id.
func Decode(packet codec.Packet) (int32, error) {
	if packet.Header != Header {
		return 0, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, codec.Definition{codec.Int32Field})
	if err != nil {
		return 0, err
	}
	return values[0].Int32, nil
}
