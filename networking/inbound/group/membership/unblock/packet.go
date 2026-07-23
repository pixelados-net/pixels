// Package unblock contains one Nitro social-group inbound packet.
package unblock

import "github.com/niflaot/pixels/networking/codec"

// Header identifies this Nitro packet.
const Header uint16 = 2864

// Payload contains the two protocol identifiers.
type Payload struct {
	// GroupID stores the first identifier.
	GroupID int64
	// PlayerID stores the second identifier.
	PlayerID int64
}

// Decode unpacks the two protocol identifiers.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, codec.Definition{codec.Int32Field, codec.Int32Field})
	if err != nil {
		return Payload{}, err
	}
	return Payload{GroupID: int64(values[0].Int32), PlayerID: int64(values[1].Int32)}, nil
}
