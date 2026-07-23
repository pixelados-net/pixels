// Package resolve contains the moderation resolve inbound packet.
package resolve

import "github.com/niflaot/pixels/networking/codec"

// Header identifies the moderation resolve packet.
const Header uint16 = 887

// Decode validates the header-only moderation resolve packet.
func Decode(packet codec.Packet) error {
	if packet.Header != Header {
		return codec.ErrUnexpectedHeader
	}
	_, err := codec.DecodePacketExact(packet, nil)
	return err
}
