// Package unaccept contains the TRADE_UNACCEPT inbound packet.
package unaccept

import "github.com/niflaot/pixels/networking/codec"

// Header identifies TRADE_UNACCEPT.
const Header uint16 = 1444

// Decode validates the header-only packet.
func Decode(packet codec.Packet) error {
	if packet.Header != Header {
		return codec.ErrUnexpectedHeader
	}
	_, err := codec.DecodePacketExact(packet, codec.Definition{})
	return err
}
