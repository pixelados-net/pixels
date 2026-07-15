// Package userinfo contains the moderation userinfo inbound packet.
package userinfo

import "github.com/niflaot/pixels/networking/codec"

// Header identifies the moderation userinfo packet.
const Header uint16 = 3295

// Payload contains decoded moderation userinfo fields.
type Payload struct {
	// PlayerID stores the decoded wire field.
	PlayerID int32
}

// Definition describes moderation userinfo fields.
var Definition = codec.Definition{
	codec.Named("playerID", codec.Int32Field),
}

// Decode validates and decodes the moderation userinfo packet.
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
