// Package rent decodes RENTABLE_SPACE_RENT requests.
package rent

import (
	"github.com/niflaot/pixels/networking/codec"
)

// Header identifies RENTABLE_SPACE_RENT.
const Header uint16 = 2946

// Decode validates a header-only rentable-space rent request.
func Decode(packet codec.Packet) error {
	if packet.Header != Header {
		return codec.ErrUnexpectedHeader
	}
	_, err := codec.DecodePacketExact(packet, nil)
	return err
}
