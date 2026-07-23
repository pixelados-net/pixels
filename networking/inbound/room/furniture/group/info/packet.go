// Package info contains one Nitro social-group inbound packet.
package info

import "github.com/niflaot/pixels/networking/codec"

// Header identifies this Nitro packet.
const Header uint16 = 2651

// Payload contains the group-furniture context identifiers.
type Payload struct {
	// ObjectID identifies the furniture item.
	ObjectID int64
	// GroupID optionally identifies the client-projected social group.
	GroupID int64
}

// Decode accepts Comet's item-only shape and Nitro or Arcturus's optional group identifier.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, rest, err := codec.DecodePacket(packet, codec.Definition{codec.Int32Field})
	if err != nil {
		return Payload{}, err
	}
	payload := Payload{ObjectID: int64(values[0].Int32)}
	if len(rest) == 0 {
		return payload, nil
	}
	if len(rest) == 2 && rest[0] == 0 && rest[1] == 0 {
		return payload, nil
	}
	optional, trailing, err := codec.DecodePayload(nil, codec.Definition{codec.Int32Field}, rest)
	if err != nil {
		return Payload{}, err
	}
	if len(trailing) != 0 {
		return Payload{}, codec.ErrUnexpectedPayload
	}
	payload.GroupID = int64(optional[0].Int32)
	return payload, nil
}
