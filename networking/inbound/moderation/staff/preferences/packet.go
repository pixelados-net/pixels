// Package preferences contains the moderation preferences inbound packet.
package preferences

import "github.com/niflaot/pixels/networking/codec"

// Header identifies the moderation preferences packet.
const Header uint16 = 31

// Decode validates the header-only moderation preferences packet.
func Decode(packet codec.Packet) error {
	if packet.Header != Header {
		return codec.ErrUnexpectedHeader
	}
	_, err := codec.DecodePacketExact(packet, nil)
	return err
}
