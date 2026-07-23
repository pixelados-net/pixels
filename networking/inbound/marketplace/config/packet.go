// Package config contains the GET_MARKETPLACE_CONFIG inbound packet.
package config

import "github.com/niflaot/pixels/networking/codec"

// Header identifies GET_MARKETPLACE_CONFIG.
const Header uint16 = 2597

// Decode validates the header-only packet.
func Decode(packet codec.Packet) error {
	if packet.Header != Header {
		return codec.ErrUnexpectedHeader
	}
	_, err := codec.DecodePacketExact(packet, codec.Definition{})
	return err
}
