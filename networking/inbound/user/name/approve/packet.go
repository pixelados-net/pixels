// Package approve decodes APPROVE_NAME requests.
package approve

import "github.com/niflaot/pixels/networking/codec"

// Header identifies APPROVE_NAME.
const Header uint16 = 2109

// Definition describes APPROVE_NAME fields.
var Definition = codec.Definition{codec.StringField, codec.Int32Field}

// Payload contains one typed name validation request.
type Payload struct {
	// Name stores the proposed name.
	Name string
	// Type identifies pet or other client-owned validation contexts.
	Type int32
}

// Decode decodes APPROVE_NAME.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return Payload{}, err
	}
	return Payload{Name: values[0].String, Type: values[1].Int32}, nil
}
