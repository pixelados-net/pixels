// Package cansell contains the CAN_SELL_MARKETPLACE inbound packet.
package cansell

import "github.com/niflaot/pixels/networking/codec"

// Header identifies CAN_SELL_MARKETPLACE.
const Header uint16 = 848

// Decode validates the header-only packet.
func Decode(packet codec.Packet) error {
	if packet.Header != Header {
		return codec.ErrUnexpectedHeader
	}
	_, err := codec.DecodePacketExact(packet, codec.Definition{})
	return err
}
