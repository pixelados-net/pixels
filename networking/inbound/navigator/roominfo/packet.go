// Package roominfo contains the GET_GUEST_ROOM inbound packet.
package roominfo

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the GET_GUEST_ROOM packet identifier.
	Header uint16 = 2230
)

// Payload contains the unpacked GET_GUEST_ROOM fields.
type Payload struct {
	// RoomID identifies the requested room.
	RoomID int32
	// EnterRoom reports whether the client intends to enter.
	EnterRoom int32
	// ForwardRoom reports whether this is a forward flow.
	ForwardRoom int32
}

// Definition describes the GET_GUEST_ROOM payload fields.
var Definition = codec.Definition{
	codec.Named("roomId", codec.Int32Field),
	codec.Named("enterRoom", codec.Int32Field),
	codec.Named("forwardRoom", codec.Int32Field),
}

// Decode unpacks a GET_GUEST_ROOM packet payload.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}

	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return Payload{}, err
	}

	return Payload{RoomID: values[0].Int32, EnterRoom: values[1].Int32, ForwardRoom: values[2].Int32}, nil
}
