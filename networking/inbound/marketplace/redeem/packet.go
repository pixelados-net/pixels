// Package redeem contains the REDEEM_MARKETPLACE_CREDITS inbound packet.
package redeem

import "github.com/niflaot/pixels/networking/codec"

// Header identifies REDEEM_MARKETPLACE_CREDITS.
const Header uint16 = 2650

// Decode validates the header-only packet.
func Decode(packet codec.Packet) error {
	if packet.Header != Header {
		return codec.ErrUnexpectedHeader
	}
	_, err := codec.DecodePacketExact(packet, codec.Definition{})
	return err
}
