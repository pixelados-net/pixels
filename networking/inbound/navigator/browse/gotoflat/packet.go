// Package gotoflat contains the GO_TO_FLAT inbound packet.
package gotoflat

import "github.com/niflaot/pixels/networking/codec"

// Header identifies GO_TO_FLAT.
const Header uint16 = 685

// Definition describes the target room identifier.
var Definition = codec.Definition{codec.Named("roomId", codec.Int32Field)}

// Payload contains one direct Navigator room admission request.
type Payload struct {
	// RoomID identifies the requested room.
	RoomID int32
}

// Decode decodes GO_TO_FLAT.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return Payload{}, err
	}
	return Payload{RoomID: values[0].Int32}, nil
}
