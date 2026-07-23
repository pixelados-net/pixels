// Package unread contains one Nitro social-group inbound packet.
package unread

import "github.com/niflaot/pixels/networking/codec"

// Header identifies this Nitro packet.
const Header uint16 = 2908

// Decode validates the empty packet payload.
func Decode(packet codec.Packet) error {
	if packet.Header != Header {
		return codec.ErrUnexpectedHeader
	}
	if len(packet.Payload) != 0 {
		return codec.ErrUnexpectedPayload
	}
	return nil
}
