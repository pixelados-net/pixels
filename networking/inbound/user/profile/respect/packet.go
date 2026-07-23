// Package respect contains the USER_RESPECT inbound packet.
package respect

import "github.com/niflaot/pixels/networking/codec"

// Header identifies USER_RESPECT.
const Header uint16 = 2694

// Definition describes USER_RESPECT fields.
var Definition = codec.Definition{codec.Named("targetPlayerId", codec.Int32Field)}

// Payload contains decoded USER_RESPECT fields.
type Payload struct {
	// TargetPlayerID identifies the respected player.
	TargetPlayerID int32
}

// Decode decodes USER_RESPECT.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return Payload{}, err
	}
	return Payload{TargetPlayerID: values[0].Int32}, nil
}
