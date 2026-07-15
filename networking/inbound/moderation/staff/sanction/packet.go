// Package sanction contains the moderation sanction inbound packet.
package sanction

import "github.com/niflaot/pixels/networking/codec"

// Header identifies the moderation sanction packet.
const Header uint16 = 1392

// Payload contains decoded moderation sanction fields.
type Payload struct {
	// PlayerID stores the decoded wire field.
	PlayerID int32
	// TopicID stores the decoded wire field.
	TopicID int32
	// SanctionID stores the decoded wire field.
	SanctionID int32
}

// Definition describes moderation sanction fields.
var Definition = codec.Definition{
	codec.Named("playerID", codec.Int32Field),
	codec.Named("topicID", codec.Int32Field),
	codec.Named("sanctionID", codec.Int32Field),
}

// Decode validates and decodes the moderation sanction packet.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return Payload{}, err
	}
	return Payload{
		PlayerID:   values[0].Int32,
		TopicID:    values[1].Int32,
		SanctionID: values[2].Int32,
	}, nil
}
