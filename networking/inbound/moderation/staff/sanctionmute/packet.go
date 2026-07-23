// Package sanctionmute contains the moderation sanctionmute inbound packet.
package sanctionmute

import "github.com/niflaot/pixels/networking/codec"

// Header identifies the moderation sanctionmute packet.
const Header uint16 = 1945

// Payload contains decoded moderation sanctionmute fields.
type Payload struct {
	// PlayerID stores the decoded wire field.
	PlayerID int32
	// Message stores the decoded wire field.
	Message string
	// TopicID identifies the selected call-for-help topic.
	TopicID int32
	// IssueID stores the decoded wire field.
	IssueID int32
}

// Definition describes moderation sanctionmute fields.
var Definition = codec.Definition{
	codec.Named("playerID", codec.Int32Field),
	codec.Named("message", codec.StringField),
	codec.Named("topicID", codec.Int32Field),
	codec.Optional(codec.Named("issueID", codec.Int32Field)),
}

// Decode validates and decodes the moderation sanctionmute packet.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return Payload{}, err
	}
	payload := Payload{
		PlayerID: values[0].Int32,
		Message:  values[1].String,
		TopicID:  values[2].Int32,
	}
	if len(values) > 3 {
		payload.IssueID = values[3].Int32
	}
	return payload, nil
}
