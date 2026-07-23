// Package cancel decodes RENTABLE_SPACE_CANCEL_RENT requests.
package cancel

import (
	"github.com/niflaot/pixels/networking/codec"
)

// Header identifies RENTABLE_SPACE_CANCEL_RENT.
const Header uint16 = 1667

// Decode validates a header-only rentable-space cancellation request.
func Decode(packet codec.Packet) error {
	if packet.Header != Header {
		return codec.ErrUnexpectedHeader
	}
	_, err := codec.DecodePacketExact(packet, nil)
	return err
}
