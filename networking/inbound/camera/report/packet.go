// Package report contains CALL_FOR_HELP_FROM_PHOTO.
package report

import "github.com/niflaot/pixels/networking/codec"

// Header identifies CALL_FOR_HELP_FROM_PHOTO.
const Header uint16 = 2492

// Payload stores one conditional photo report request.
type Payload struct {
	// ExtraDataID stores optional client evidence.
	ExtraDataID string
	// RoomID identifies the reported room.
	RoomID int32
	// ReportedPlayerID stores the untrusted client target.
	ReportedPlayerID int32
	// TopicID identifies the moderation topic.
	TopicID int32
	// ItemID identifies the placed photo furniture.
	ItemID int32
}

// Decode decodes the optional evidence string and fixed report suffix.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	definition := codec.Definition{codec.StringField, codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field}
	values, rest, err := codec.DecodePacket(packet, definition)
	if err != nil {
		return Payload{}, err
	}
	if len(rest) != 0 || values[1].Int32 <= 0 || values[3].Int32 <= 0 || values[4].Int32 <= 0 {
		return Payload{}, codec.ErrInvalidField
	}
	return Payload{ExtraDataID: values[0].String, RoomID: values[1].Int32, ReportedPlayerID: values[2].Int32, TopicID: values[3].Int32, ItemID: values[4].Int32}, nil
}
