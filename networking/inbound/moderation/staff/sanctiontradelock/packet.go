// Package sanctiontradelock contains the moderation sanctiontradelock inbound packet.
package sanctiontradelock

import "github.com/niflaot/pixels/networking/codec"

// Header identifies the moderation sanctiontradelock packet.
const Header uint16 = 3742

// Payload contains decoded moderation sanctiontradelock fields.
type Payload struct {
	// PlayerID stores the decoded wire field.
	PlayerID int32
	// Message stores the decoded wire field.
	Message string
	// Minutes stores Nitro's selected trade-lock duration.
	Minutes int32
	// TopicID stores the decoded wire field.
	TopicID int32
	// IssueID stores the decoded wire field.
	IssueID int32
}

// Definition describes moderation sanctiontradelock fields.
var Definition = codec.Definition{
	codec.Named("playerID", codec.Int32Field),
	codec.Named("message", codec.StringField),
	codec.Named("minutes", codec.Int32Field),
	codec.Named("topicID", codec.Int32Field),
	codec.Optional(codec.Named("issueID", codec.Int32Field)),
}

// Decode validates and decodes the moderation sanctiontradelock packet.
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
		Minutes:  values[2].Int32,
		TopicID:  values[3].Int32,
	}
	if len(values) > 4 {
		payload.IssueID = values[4].Int32
	}
	return payload, nil
}
