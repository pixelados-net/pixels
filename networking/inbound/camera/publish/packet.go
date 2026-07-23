// Package publish contains PUBLISH_PHOTO.
package publish

import "github.com/niflaot/pixels/networking/codec"

// Header identifies PUBLISH_PHOTO.
const Header uint16 = 2068

// Decode validates the header-only request.
func Decode(packet codec.Packet) error {
	if packet.Header != Header {
		return codec.ErrUnexpectedHeader
	}
	_, err := codec.DecodePacketExact(packet, nil)
	return err
}
