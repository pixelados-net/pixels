// Package tokens contains the BUY_MARKETPLACE_TOKENS inbound packet.
package tokens

import "github.com/niflaot/pixels/networking/codec"

// Header identifies BUY_MARKETPLACE_TOKENS.
const Header uint16 = 1866

// Decode validates the header-only packet.
func Decode(packet codec.Packet) error {
	if packet.Header != Header {
		return codec.ErrUnexpectedHeader
	}
	_, err := codec.DecodePacketExact(packet, codec.Definition{})
	return err
}
