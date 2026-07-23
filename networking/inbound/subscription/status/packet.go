// Package status contains the USER_SUBSCRIPTION inbound packet.
package status

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the USER_SUBSCRIPTION packet identifier.
	Header uint16 = 3166
)

// Payload contains the requested club product.
type Payload struct {
	// ProductName stores the club product identifier.
	ProductName string
}

// Definition describes the required packet fields.
var Definition = codec.Definition{codec.StringField}

// Decode decodes a USER_SUBSCRIPTION packet.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return Payload{}, err
	}
	return Payload{ProductName: values[0].String}, nil
}
