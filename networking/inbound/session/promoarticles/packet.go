// Package promoarticles decodes the DESKTOP_NEWS inbound request.
package promoarticles

import "github.com/niflaot/pixels/networking/codec"

// Header identifies DESKTOP_NEWS.
const Header uint16 = 1827

// Decode validates one header-only DESKTOP_NEWS request.
func Decode(packet codec.Packet) error {
	if packet.Header != Header {
		return codec.ErrUnexpectedHeader
	}
	_, err := codec.DecodePacketExact(packet, codec.Definition{})
	return err
}
