// Package cancel contains the moderation cancel inbound packet.
package cancel

import "github.com/niflaot/pixels/networking/codec"

// Header identifies the moderation cancel packet.
const Header uint16 = 291

// Decode validates the header-only moderation cancel packet.
func Decode(packet codec.Packet) error {
	if packet.Header != Header {
		return codec.ErrUnexpectedHeader
	}
	_, err := codec.DecodePacketExact(packet, nil)
	return err
}
