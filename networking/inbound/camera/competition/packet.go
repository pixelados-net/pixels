// Package competition contains PHOTO_COMPETITION.
package competition

import "github.com/niflaot/pixels/networking/codec"

// Header identifies PHOTO_COMPETITION.
const Header uint16 = 3959

// Decode validates the header-only compatibility request.
func Decode(packet codec.Packet) error {
	if packet.Header != Header {
		return codec.ErrUnexpectedHeader
	}
	_, err := codec.DecodePacketExact(packet, nil)
	return err
}
