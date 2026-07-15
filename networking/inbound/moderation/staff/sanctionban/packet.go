// Package sanctionban contains the moderation sanctionban inbound packet.
package sanctionban

import "github.com/niflaot/pixels/networking/codec"

// Header identifies the moderation sanctionban packet.
const Header uint16 = 2766

// Payload contains decoded moderation sanctionban fields.
type Payload struct {
	// PlayerID stores the decoded wire field.
	PlayerID int32
	// Message stores the decoded wire field.
	Message string
	// Hours stores the decoded wire field.
	Hours int32
	// TopicID stores the decoded wire field.
	TopicID int32
	// Permanent stores the decoded wire field.
	Permanent bool
	// IssueID stores the decoded wire field.
	IssueID int32
}

// Definition describes moderation sanctionban fields.
var Definition = codec.Definition{
	codec.Named("playerID", codec.Int32Field),
	codec.Named("message", codec.StringField),
	codec.Named("hours", codec.Int32Field),
	codec.Named("topicID", codec.Int32Field),
	codec.Named("permanent", codec.BooleanField),
	codec.Optional(codec.Named("issueID", codec.Int32Field)),
}

// Decode validates and decodes the moderation sanctionban packet.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return Payload{}, err
	}
	payload := Payload{
		PlayerID:  values[0].Int32,
		Message:   values[1].String,
		Hours:     values[2].Int32,
		TopicID:   values[3].Int32,
		Permanent: values[4].Boolean,
	}
	if len(values) > 5 {
		payload.IssueID = values[5].Int32
	}
	return payload, nil
}
