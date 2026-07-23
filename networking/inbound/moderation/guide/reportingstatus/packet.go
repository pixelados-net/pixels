// Package reportingstatus contains the moderation reportingstatus inbound packet.
package reportingstatus

import "github.com/niflaot/pixels/networking/codec"

// Header identifies the moderation reportingstatus packet.
const Header uint16 = 3786

// Decode validates the header-only moderation reportingstatus packet.
func Decode(packet codec.Packet) error {
	if packet.Header != Header {
		return codec.ErrUnexpectedHeader
	}
	_, err := codec.DecodePacketExact(packet, nil)
	return err
}
