// Package invite contains the moderation invite inbound packet.
package invite

import "github.com/niflaot/pixels/networking/codec"

// Header identifies the moderation invite packet.
const Header uint16 = 234

// Decode validates the header-only moderation invite packet.
func Decode(packet codec.Packet) error {
	if packet.Header != Header {
		return codec.ErrUnexpectedHeader
	}
	_, err := codec.DecodePacketExact(packet, nil)
	return err
}
