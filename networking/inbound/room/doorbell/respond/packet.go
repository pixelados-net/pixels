// Package respond contains the ROOM_DOORBELL inbound packet.
package respond

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the ROOM_DOORBELL packet identifier.
	Header uint16 = 1644
)

// Payload contains unpacked ROOM_DOORBELL fields.
type Payload struct {
	// Username identifies the waiting player.
	Username string
	// Accepted reports whether room entry was approved.
	Accepted bool
}

// Definition describes ROOM_DOORBELL fields.
var Definition = codec.Definition{
	codec.Named("username", codec.StringField),
	codec.Named("accepted", codec.BooleanField),
}

// Decode unpacks a ROOM_DOORBELL packet.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return Payload{}, err
	}

	return Payload{Username: values[0].String, Accepted: values[1].Boolean}, nil
}
