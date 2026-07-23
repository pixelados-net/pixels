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
	// TopicID stores the decoded wire field.
	TopicID int32
	// ActionIndex identifies Nitro's selected ban option.
	ActionIndex int32
	// AvatarOnly reports Nitro's avatar-only long-ban option.
	AvatarOnly bool
	// IssueID stores the decoded wire field.
	IssueID int32
}

// Definition describes moderation sanctionban fields.
var Definition = codec.Definition{
	codec.Named("playerID", codec.Int32Field),
	codec.Named("message", codec.StringField),
	codec.Named("topicID", codec.Int32Field),
	codec.Named("actionIndex", codec.Int32Field),
	codec.Named("avatarOnly", codec.BooleanField),
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
		PlayerID:    values[0].Int32,
		Message:     values[1].String,
		TopicID:     values[2].Int32,
		ActionIndex: values[3].Int32,
		AvatarOnly:  values[4].Boolean,
	}
	if len(values) > 5 {
		payload.IssueID = values[5].Int32
	}
	return payload, nil
}
