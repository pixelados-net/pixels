// Package status contains the retired EMAIL_GET_STATUS inbound packet.
package status

import "github.com/niflaot/pixels/networking/codec"

// Header identifies EMAIL_GET_STATUS.
const Header uint16 = 2557

// Definition describes the empty EMAIL_GET_STATUS payload.
var Definition = codec.Definition{}

// Payload contains the decoded EMAIL_GET_STATUS fields.
type Payload struct{}

// Decode validates an EMAIL_GET_STATUS packet.
//
// Deprecated: email is CMS-owned and intentionally has no hotel behavior.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	_, err := codec.DecodePacketExact(packet, Definition)
	return Payload{}, err
}
