// Package userchatlog contains the moderation userchatlog inbound packet.
package userchatlog

import "github.com/niflaot/pixels/networking/codec"

// Header identifies the moderation userchatlog packet.
const Header uint16 = 1391

// Payload contains decoded moderation userchatlog fields.
type Payload struct {
	// PlayerID stores the decoded wire field.
	PlayerID int32
}

// Definition describes moderation userchatlog fields.
var Definition = codec.Definition{
	codec.Named("playerID", codec.Int32Field),
}

// Decode validates and decodes the moderation userchatlog packet.
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
