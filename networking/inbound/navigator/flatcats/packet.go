// Package flatcats contains the GET_USER_FLAT_CATS inbound packet.
package flatcats

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the GET_USER_FLAT_CATS packet identifier.
	Header uint16 = 3027
)

// Payload contains the unpacked GET_USER_FLAT_CATS fields.
type Payload struct{}

// Definition describes the GET_USER_FLAT_CATS payload fields.
var Definition = codec.Definition{}

// Decode unpacks a GET_USER_FLAT_CATS packet payload.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	if _, err := codec.DecodePacketExact(packet, Definition); err != nil {
		return Payload{}, err
	}
	return Payload{}, nil
}
