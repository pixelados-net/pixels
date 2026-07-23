// Package start contains the UNIT_TYPING inbound packet.
package start

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header identifies UNIT_TYPING.
	Header uint16 = 1597
)

// Definition describes the empty typing-start payload.
var Definition = codec.Definition{}

// Decode validates a UNIT_TYPING packet.
func Decode(packet codec.Packet) error {
	if packet.Header != Header {
		return codec.ErrUnexpectedHeader
	}
	_, err := codec.DecodePacketExact(packet, Definition)
	return err
}
