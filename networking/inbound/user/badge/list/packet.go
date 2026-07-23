// Package list decodes player badge inventory requests.
package list

import "github.com/niflaot/pixels/networking/codec"

// Header identifies USER_BADGES requests.
const Header uint16 = 2769

// Decode validates a USER_BADGES request.
func Decode(packet codec.Packet) error {
	if packet.Header != Header {
		return codec.ErrUnexpectedHeader
	}
	if len(packet.Payload) != 0 {
		return codec.ErrUnexpectedPayload
	}
	return nil
}
