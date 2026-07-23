// Package categorycounts contains the GET_CATEGORIES_WITH_USER_COUNT inbound packet.
package categorycounts

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the GET_CATEGORIES_WITH_USER_COUNT packet identifier.
	Header uint16 = 3782
)

// Payload contains the unpacked GET_CATEGORIES_WITH_USER_COUNT fields.
type Payload struct{}

// Definition describes the GET_CATEGORIES_WITH_USER_COUNT payload fields.
var Definition = codec.Definition{}

// Decode unpacks a GET_CATEGORIES_WITH_USER_COUNT packet payload.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	if _, err := codec.DecodePacketExact(packet, Definition); err != nil {
		return Payload{}, err
	}
	return Payload{}, nil
}
