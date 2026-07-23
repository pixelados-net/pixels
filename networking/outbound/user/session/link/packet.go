// Package link contains the IN_CLIENT_LINK outbound packet.
package link

import "github.com/niflaot/pixels/networking/codec"

// Header identifies IN_CLIENT_LINK.
const Header uint16 = 2023

// Definition describes IN_CLIENT_LINK fields.
var Definition = codec.Definition{codec.Named("link", codec.StringField)}

// Encode creates an IN_CLIENT_LINK packet.
func Encode(link string) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.String(link))
}
