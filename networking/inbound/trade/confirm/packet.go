// Package confirm contains the TRADE_CONFIRM inbound packet.
package confirm

import "github.com/niflaot/pixels/networking/codec"

// Header identifies TRADE_CONFIRM.
const Header uint16 = 2760

// Decode validates the header-only packet.
func Decode(packet codec.Packet) error {
	if packet.Header != Header {
		return codec.ErrUnexpectedHeader
	}
	_, err := codec.DecodePacketExact(packet, codec.Definition{})
	return err
}
