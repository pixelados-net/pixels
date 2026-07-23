// Package findroom contains FIND_NEW_FRIENDS.
package findroom

import "github.com/niflaot/pixels/networking/codec"

// Header identifies FIND_NEW_FRIENDS.
const Header uint16 = 516

// Decode validates FIND_NEW_FRIENDS.
func Decode(packet codec.Packet) error {
	if packet.Header != Header {
		return codec.ErrUnexpectedHeader
	}
	if len(packet.Payload) != 0 {
		return codec.ErrUnexpectedPayload
	}
	return nil
}
