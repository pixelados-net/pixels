// Package elementpointer encodes the retired NOTIFICATION_ELEMENT_POINTER packet.
package elementpointer

import "github.com/niflaot/pixels/networking/codec"

// Header identifies NOTIFICATION_ELEMENT_POINTER.
const Header uint16 = 1787

// Definition describes the header-only legacy packet.
var Definition = codec.Definition{}

// Encode creates one compatibility packet.
func Encode() (codec.Packet, error) {
	return codec.NewPacket(Header, Definition)
}
