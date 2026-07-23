// Package pending contains the moderation pending inbound packet.
package pending

import "github.com/niflaot/pixels/networking/codec"

// Header identifies the moderation pending packet.
const Header uint16 = 3267

// Decode validates the header-only moderation pending packet.
func Decode(packet codec.Packet) error {
	if packet.Header != Header {
		return codec.ErrUnexpectedHeader
	}
	_, err := codec.DecodePacketExact(packet, nil)
	return err
}
