// Package promoarticles encodes the DESKTOP_NEWS outbound response.
package promoarticles

import "github.com/niflaot/pixels/networking/codec"

// Header identifies DESKTOP_NEWS.
const Header uint16 = 286

// Encode creates an explicit empty promo-article list.
func Encode() (codec.Packet, error) {
	return codec.NewPacket(Header, codec.Definition{codec.Int32Field}, codec.Int32(0))
}
