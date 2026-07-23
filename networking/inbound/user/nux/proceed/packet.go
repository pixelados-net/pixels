// Package proceed contains the retired NEW_USER_EXPERIENCE_SCRIPT_PROCEED packet.
package proceed

import "github.com/niflaot/pixels/networking/codec"

// Header identifies NEW_USER_EXPERIENCE_SCRIPT_PROCEED.
const Header uint16 = 1299

// Definition describes the empty NUX proceed payload.
var Definition = codec.Definition{}

// Payload contains decoded NUX proceed fields.
type Payload struct{}

// Decode validates a NUX proceed packet.
//
// Deprecated: the legacy NUX journey is intentionally retired.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	_, err := codec.DecodePacketExact(packet, Definition)
	return Payload{}, err
}
