// Package search contains the NAVIGATOR_SEARCH inbound packet.
package search

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the NAVIGATOR_SEARCH packet identifier.
	Header uint16 = 249
)

// Payload contains the unpacked NAVIGATOR_SEARCH fields.
type Payload struct {
	// Code identifies the navigator search context.
	Code string
	// Data stores the query or filter string.
	Data string
}

// Definition describes the NAVIGATOR_SEARCH payload fields.
var Definition = codec.Definition{
	codec.Named("code", codec.StringField),
	codec.Named("data", codec.StringField),
}

// Decode unpacks a NAVIGATOR_SEARCH packet payload.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}

	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return Payload{}, err
	}

	return Payload{Code: values[0].String, Data: values[1].String}, nil
}
