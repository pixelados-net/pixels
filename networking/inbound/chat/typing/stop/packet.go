// Package stop contains the UNIT_TYPING_STOP inbound packet.
package stop

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header identifies UNIT_TYPING_STOP.
	Header uint16 = 1474
)

// Definition describes the empty typing-stop payload.
var Definition = codec.Definition{}

// Decode validates a UNIT_TYPING_STOP packet.
func Decode(packet codec.Packet) error {
	if packet.Header != Header {
		return codec.ErrUnexpectedHeader
	}
	_, err := codec.DecodePacketExact(packet, Definition)
	return err
}
