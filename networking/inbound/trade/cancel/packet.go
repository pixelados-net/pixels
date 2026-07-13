// Package cancel contains the TRADE_CANCEL_CONFIRMATION inbound packet.
package cancel

import "github.com/niflaot/pixels/networking/codec"

// Header identifies TRADE_CANCEL_CONFIRMATION.
const Header uint16 = 2341

// Decode validates the header-only packet.
func Decode(packet codec.Packet) error {
	if packet.Header != Header {
		return codec.ErrUnexpectedHeader
	}
	_, err := codec.DecodePacketExact(packet, codec.Definition{})
	return err
}
