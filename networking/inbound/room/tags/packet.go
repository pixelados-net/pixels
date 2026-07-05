// Package tags contains room tag inbound packets.
package tags

import "github.com/niflaot/pixels/networking/codec"

const (
	// SessionHeader is the SET_ROOM_SESSION_TAGS packet identifier.
	SessionHeader uint16 = 3305
	// PopularHeader is the GET_POPULAR_ROOM_TAGS packet identifier.
	PopularHeader uint16 = 826
)

// Payload contains the unpacked room tag fields.
type Payload struct {
	// Header stores which room tag packet was received.
	Header uint16
}

// Definition describes the room tag payload fields.
var Definition = codec.Definition{}

// Decode unpacks a room tag packet payload.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != SessionHeader && packet.Header != PopularHeader {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	if _, err := codec.DecodePacketExact(packet, Definition); err != nil {
		return Payload{}, err
	}
	return Payload{Header: packet.Header}, nil
}
