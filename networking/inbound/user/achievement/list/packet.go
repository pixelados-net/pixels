// Package list decodes the ACHIEVEMENT_LIST inbound request.
package list

import "github.com/niflaot/pixels/networking/codec"

// Header identifies ACHIEVEMENT_LIST.
const Header uint16 = 219

// Definition describes the header-only request.
var Definition = codec.Definition{}

// Decode validates one ACHIEVEMENT_LIST request.
func Decode(packet codec.Packet) error {
	if packet.Header != Header {
		return codec.ErrUnexpectedHeader
	}
	_, err := codec.DecodePacketExact(packet, Definition)
	return err
}
