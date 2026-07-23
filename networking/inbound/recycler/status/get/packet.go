// Package get contains the RECYCLER_STATUS inbound packet.
package get

import "github.com/niflaot/pixels/networking/codec"

// Header identifies RECYCLER_STATUS.
const Header uint16 = 1342

// Decode validates one empty recycler status request.
func Decode(packet codec.Packet) error {
	if packet.Header != Header {
		return codec.ErrUnexpectedHeader
	}
	if len(packet.Payload) != 0 {
		return codec.ErrUnexpectedPayload
	}
	return nil
}
