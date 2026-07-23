// Package alertevent contains the moderation alertevent inbound packet.
package alertevent

import "github.com/niflaot/pixels/networking/codec"

// Header identifies the moderation alertevent packet.
const Header uint16 = 1840

// Payload contains decoded moderation alertevent fields.
type Payload struct {
	// PlayerID stores the decoded wire field.
	PlayerID int32
	// Message stores the decoded wire field.
	Message string
	// UnusedA stores the decoded wire field.
	UnusedA string
	// UnusedB stores the decoded wire field.
	UnusedB string
	// TopicID stores the decoded wire field.
	TopicID int32
	// IssueID stores the decoded wire field.
	IssueID int32
}

// Definition describes moderation alertevent fields.
var Definition = codec.Definition{
	codec.Named("playerID", codec.Int32Field),
	codec.Named("message", codec.StringField),
	codec.Named("unusedA", codec.StringField),
	codec.Named("unusedB", codec.StringField),
	codec.Named("topicID", codec.Int32Field),
	codec.Optional(codec.Named("issueID", codec.Int32Field)),
}

// Decode validates and decodes the moderation alertevent packet.
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
		UnusedA:  values[2].String,
		UnusedB:  values[3].String,
		TopicID:  values[4].Int32,
	}
	if len(values) > 5 {
		payload.IssueID = values[5].Int32
	}
	return payload, nil
}
