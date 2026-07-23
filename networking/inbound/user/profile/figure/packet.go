// Package figure contains the USER_FIGURE inbound packet.
package figure

import "github.com/niflaot/pixels/networking/codec"

// Header identifies USER_FIGURE.
const Header uint16 = 2730

// Definition describes USER_FIGURE fields.
var Definition = codec.Definition{codec.Named("gender", codec.StringField), codec.Named("figure", codec.StringField)}

// Payload contains decoded USER_FIGURE fields.
type Payload struct {
	// Gender stores the requested gender code.
	Gender string
	// Figure stores the requested avatar figure.
	Figure string
}

// Decode decodes USER_FIGURE.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return Payload{}, err
	}
	return Payload{Gender: values[0].String, Figure: values[1].String}, nil
}
