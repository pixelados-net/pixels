// Package selfie contains the moderation selfie inbound packet.
package selfie

import "github.com/niflaot/pixels/networking/codec"

// Header identifies the moderation selfie packet.
const Header uint16 = 2755

// Payload contains decoded moderation selfie fields.
type Payload struct {
	// Message stores the decoded wire field.
	Message string
	// RoomID stores the decoded wire field.
	RoomID int32
	// ReportedPlayerID stores the decoded wire field.
	ReportedPlayerID int32
	// ExtraData stores the decoded wire field.
	ExtraData string
	// TopicID stores the decoded wire field.
	TopicID int32
}

// Definition describes moderation selfie fields.
var Definition = codec.Definition{
	codec.Named("message", codec.StringField),
	codec.Named("roomID", codec.Int32Field),
	codec.Named("reportedPlayerID", codec.Int32Field),
	codec.Named("extraData", codec.StringField),
	codec.Named("topicID", codec.Int32Field),
}

// Decode validates and decodes the moderation selfie packet.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return Payload{}, err
	}
	return Payload{
		Message:          values[0].String,
		RoomID:           values[1].Int32,
		ReportedPlayerID: values[2].Int32,
		ExtraData:        values[3].String,
		TopicID:          values[4].Int32,
	}, nil
}
