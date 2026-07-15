// Package detach contains the moderation detach inbound packet.
package detach

import "github.com/niflaot/pixels/networking/codec"

// Header identifies the moderation detach packet.
const Header uint16 = 2501

// Decode validates the header-only moderation detach packet.
func Decode(packet codec.Packet) error {
	if packet.Header != Header {
		return codec.ErrUnexpectedHeader
	}
	_, err := codec.DecodePacketExact(packet, nil)
	return err
}
