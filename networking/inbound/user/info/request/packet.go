// Package request contains the USER_INFO inbound request.
package request

import "github.com/niflaot/pixels/networking/codec"

// Header identifies USER_INFO.
const Header uint16 = 357

// Definition describes the empty USER_INFO request.
var Definition = codec.Definition{}

// Payload contains decoded USER_INFO request fields.
type Payload struct{}

// Decode validates a USER_INFO request.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	_, err := codec.DecodePacketExact(packet, Definition)
	return Payload{}, err
}
