// Package enter contains the ROOM_ENTER inbound packet.
package enter

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the ROOM_ENTER packet identifier.
	Header uint16 = 2312
)

// Payload contains the unpacked ROOM_ENTER fields.
type Payload struct {
	// FlatID identifies the room to enter.
	FlatID int32

	// Password stores the optional room password.
	Password string
}

// Definition describes the ROOM_ENTER payload fields.
var Definition = codec.Definition{
	codec.Named("flatId", codec.Int32Field),
	codec.Named("password", codec.StringField),
}

// Decode unpacks a ROOM_ENTER packet payload.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return Payload{}, err
	}
	return Payload{FlatID: values[0].Int32, Password: values[1].String}, nil
}
