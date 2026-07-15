// Package deletepending contains the moderation deletepending inbound packet.
package deletepending

import "github.com/niflaot/pixels/networking/codec"

// Header identifies the moderation deletepending packet.
const Header uint16 = 3605

// Decode validates the header-only moderation deletepending packet.
func Decode(packet codec.Packet) error {
	if packet.Header != Header {
		return codec.ErrUnexpectedHeader
	}
	_, err := codec.DecodePacketExact(packet, nil)
	return err
}
