// Package status decodes RENTABLE_SPACE_STATUS requests.
package status

import (
	"github.com/niflaot/pixels/networking/codec"
)

// Header identifies RENTABLE_SPACE_STATUS.
const Header uint16 = 872

// Decode validates a header-only rentable-space status request.
func Decode(packet codec.Packet) error {
	if packet.Header != Header {
		return codec.ErrUnexpectedHeader
	}
	_, err := codec.DecodePacketExact(packet, nil)
	return err
}
