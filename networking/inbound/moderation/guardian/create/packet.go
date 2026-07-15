// Package create contains the moderation create inbound packet.
package create

import "github.com/niflaot/pixels/networking/codec"

// Header identifies the moderation create packet.
const Header uint16 = 3060

// Payload contains decoded moderation create fields.
type Payload struct {
	// ReportedPlayerID stores the decoded wire field.
	ReportedPlayerID int32
	// RoomID stores the decoded wire field.
	RoomID int32
}

// Definition describes moderation create fields.
var Definition = codec.Definition{
	codec.Named("reportedPlayerID", codec.Int32Field),
	codec.Named("roomID", codec.Int32Field),
}

// Decode validates and decodes the moderation create packet.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return Payload{}, err
	}
	return Payload{
		ReportedPlayerID: values[0].Int32,
		RoomID:           values[1].Int32,
	}, nil
}
