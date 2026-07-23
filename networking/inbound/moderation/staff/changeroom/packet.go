// Package changeroom contains the moderation changeroom inbound packet.
package changeroom

import "github.com/niflaot/pixels/networking/codec"

// Header identifies the moderation changeroom packet.
const Header uint16 = 3260

// Payload contains decoded moderation changeroom fields.
type Payload struct {
	// RoomID stores the decoded wire field.
	RoomID int32
	// LockDoor stores the decoded wire field.
	LockDoor int32
	// ChangeTitle stores the decoded wire field.
	ChangeTitle int32
	// KickUsers stores the decoded wire field.
	KickUsers int32
}

// Definition describes moderation changeroom fields.
var Definition = codec.Definition{
	codec.Named("roomID", codec.Int32Field),
	codec.Named("lockDoor", codec.Int32Field),
	codec.Named("changeTitle", codec.Int32Field),
	codec.Named("kickUsers", codec.Int32Field),
}

// Decode validates and decodes the moderation changeroom packet.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return Payload{}, err
	}
	return Payload{
		RoomID:      values[0].Int32,
		LockDoor:    values[1].Int32,
		ChangeTitle: values[2].Int32,
		KickUsers:   values[3].Int32,
	}, nil
}
