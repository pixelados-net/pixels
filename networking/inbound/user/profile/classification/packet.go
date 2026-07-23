// Package classification contains the compatibility USER_CLASSIFICATION inbound packet.
package classification

import "github.com/niflaot/pixels/networking/codec"

// Header identifies USER_CLASSIFICATION.
const Header uint16 = 2285

// Definition describes USER_CLASSIFICATION fields.
var Definition = codec.Definition{codec.Named("classType", codec.StringField)}

// Payload contains decoded USER_CLASSIFICATION fields.
type Payload struct {
	// ClassType stores the ignored compatibility classification.
	ClassType string
}

// Decode decodes USER_CLASSIFICATION without granting authority.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return Payload{}, err
	}
	return Payload{ClassType: values[0].String}, nil
}
