// Package accept contains the TRADE_ACCEPT inbound packet.
package accept

import "github.com/niflaot/pixels/networking/codec"

// Header identifies TRADE_ACCEPT.
const Header uint16 = 3863

// Decode validates the header-only packet.
func Decode(packet codec.Packet) error {
	if packet.Header != Header {
		return codec.ErrUnexpectedHeader
	}
	_, err := codec.DecodePacketExact(packet, codec.Definition{})
	return err
}
