// Package entrytile contains the GET_ROOM_ENTRY_TILE inbound packet.
package entrytile

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the GET_ROOM_ENTRY_TILE packet identifier.
	Header uint16 = 3559
)

// Payload contains the unpacked GET_ROOM_ENTRY_TILE fields.
type Payload struct{}

// Definition describes the GET_ROOM_ENTRY_TILE payload fields.
var Definition = codec.Definition{}

// Decode unpacks a GET_ROOM_ENTRY_TILE packet payload.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	if _, err := codec.DecodePacketExact(packet, Definition); err != nil {
		return Payload{}, err
	}

	return Payload{}, nil
}
