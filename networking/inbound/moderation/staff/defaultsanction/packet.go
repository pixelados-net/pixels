// Package defaultsanction contains the moderation defaultsanction inbound packet.
package defaultsanction

import "github.com/niflaot/pixels/networking/codec"

// Header identifies the moderation defaultsanction packet.
const Header uint16 = 1681

// Payload contains decoded moderation defaultsanction fields.
type Payload struct {
	// PlayerID stores the decoded wire field.
	PlayerID int32
	// TopicID stores the decoded wire field.
	TopicID int32
	// Reason stores the decoded wire field.
	Reason string
	// IssueID stores the decoded wire field.
	IssueID int32
}

// Definition describes moderation defaultsanction fields.
var Definition = codec.Definition{
	codec.Named("playerID", codec.Int32Field),
	codec.Named("topicID", codec.Int32Field),
	codec.Named("reason", codec.StringField),
	codec.Optional(codec.Named("issueID", codec.Int32Field)),
}

// Decode validates and decodes the moderation defaultsanction packet.
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
		TopicID:  values[1].Int32,
		Reason:   values[2].String,
	}
	if len(values) > 3 {
		payload.IssueID = values[3].Int32
	}
	return payload, nil
}
