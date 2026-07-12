// Package relation contains SET_RELATIONSHIP_STATUS.
package relation

import "github.com/niflaot/pixels/networking/codec"

// Header identifies SET_RELATIONSHIP_STATUS.
const Header uint16 = 3768

// Payload contains relationship fields.
type Payload struct {
	// PlayerID identifies the friend.
	PlayerID int64
	// Relation stores Nitro's relationship marker.
	Relation int16
}

// Decode unpacks SET_RELATIONSHIP_STATUS.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, codec.Definition{codec.Named("playerId", codec.Int32Field), codec.Named("relation", codec.Int32Field)})
	if err != nil {
		return Payload{}, err
	}
	return Payload{PlayerID: int64(values[0].Int32), Relation: int16(values[1].Int32)}, nil
}
