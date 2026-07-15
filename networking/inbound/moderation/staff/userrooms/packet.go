// Package userrooms contains the moderation userrooms inbound packet.
package userrooms

import "github.com/niflaot/pixels/networking/codec"

// Header identifies the moderation userrooms packet.
const Header uint16 = 3526

// Payload contains decoded moderation userrooms fields.
type Payload struct {
	// PlayerID stores the decoded wire field.
	PlayerID int32
}

// Definition describes moderation userrooms fields.
var Definition = codec.Definition{
	codec.Named("playerID", codec.Int32Field),
}

// Decode validates and decodes the moderation userrooms packet.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return Payload{}, err
	}
	return Payload{
		PlayerID: values[0].Int32,
	}, nil
}
