// Package get contains the compatibility RECYCLER_PRIZES inbound packet.
package get

import "github.com/niflaot/pixels/networking/codec"

// Header identifies RECYCLER_PRIZES.
const Header uint16 = 398

// Decode validates one empty recycler prize request.
func Decode(packet codec.Packet) error {
	if packet.Header != Header {
		return codec.ErrUnexpectedHeader
	}
	if len(packet.Payload) != 0 {
		return codec.ErrUnexpectedPayload
	}
	return nil
}
