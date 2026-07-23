// Package aliases decodes the FURNITURE_ALIASES inbound request.
package aliases

import "github.com/niflaot/pixels/networking/codec"

// Header identifies FURNITURE_ALIASES.
const Header uint16 = 3898

// Decode validates one header-only FURNITURE_ALIASES request.
func Decode(packet codec.Packet) error {
	if packet.Header != Header {
		return codec.ErrUnexpectedHeader
	}
	_, err := codec.DecodePacketExact(packet, codec.Definition{})
	return err
}
