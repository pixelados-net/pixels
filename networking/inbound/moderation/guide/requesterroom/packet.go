// Package requesterroom contains the moderation requesterroom inbound packet.
package requesterroom

import "github.com/niflaot/pixels/networking/codec"

// Header identifies the moderation requesterroom packet.
const Header uint16 = 1052

// Decode validates the header-only moderation requesterroom packet.
func Decode(packet codec.Packet) error {
	if packet.Header != Header {
		return codec.ErrUnexpectedHeader
	}
	_, err := codec.DecodePacketExact(packet, nil)
	return err
}
