// Package request contains the GET_CATALOG_INDEX inbound packet.
package request

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the GET_CATALOG_INDEX packet identifier.
	Header uint16 = 1195
)

// Payload contains the unpacked GET_CATALOG_INDEX fields.
type Payload struct {
	// Mode identifies the requested catalog mode.
	Mode string
}

// Definition describes the GET_CATALOG_INDEX payload fields.
var Definition = codec.Definition{codec.Named("mode", codec.StringField)}

// Decode unpacks a GET_CATALOG_INDEX packet payload.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return Payload{}, err
	}

	return Payload{Mode: values[0].String}, nil
}
