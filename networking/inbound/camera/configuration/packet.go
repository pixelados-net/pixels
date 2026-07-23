// Package configuration contains REQUEST_CAMERA_CONFIGURATION.
package configuration

import "github.com/niflaot/pixels/networking/codec"

// Header identifies REQUEST_CAMERA_CONFIGURATION.
const Header uint16 = 796

// Decode validates the header-only request.
func Decode(packet codec.Packet) error {
	if packet.Header != Header {
		return codec.ErrUnexpectedHeader
	}
	_, err := codec.DecodePacketExact(packet, nil)
	return err
}
