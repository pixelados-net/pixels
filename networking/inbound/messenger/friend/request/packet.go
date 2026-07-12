// Package request contains REQUEST_FRIEND.
package request

import "github.com/niflaot/pixels/networking/codec"

// Header identifies REQUEST_FRIEND.
const Header uint16 = 3157

// Payload contains unpacked REQUEST_FRIEND fields.
type Payload struct {
	// Username identifies the requested friend.
	Username string
}

// Decode unpacks REQUEST_FRIEND.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, codec.Definition{codec.Named("username", codec.StringField)})
	if err != nil {
		return Payload{}, err
	}
	return Payload{Username: values[0].String}, nil
}
