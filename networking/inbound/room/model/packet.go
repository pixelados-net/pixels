// Package model contains the ROOM_MODEL inbound packet.
package model

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the ROOM_MODEL packet identifier.
	Header uint16 = 2300
)

// Payload contains the unpacked ROOM_MODEL fields.
type Payload struct{}

// Definition describes the ROOM_MODEL payload fields.
var Definition = codec.Definition{}

// Decode unpacks a ROOM_MODEL packet payload.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	if _, err := codec.DecodePacketExact(packet, Definition); err != nil {
		return Payload{}, err
	}
	return Payload{}, nil
}
