// Package close contains the TRADE_CLOSE inbound packet.
package close

import "github.com/niflaot/pixels/networking/codec"

// Header identifies TRADE_CLOSE.
const Header uint16 = 2551

// Decode validates the header-only packet.
func Decode(packet codec.Packet) error {
	if packet.Header != Header {
		return codec.ErrUnexpectedHeader
	}
	_, err := codec.DecodePacketExact(packet, codec.Definition{})
	return err
}
