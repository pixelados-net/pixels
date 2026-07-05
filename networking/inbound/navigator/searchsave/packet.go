// Package searchsave contains the NAVIGATOR_SEARCH_SAVE inbound packet.
package searchsave

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the NAVIGATOR_SEARCH_SAVE packet identifier.
	Header uint16 = 2226
)

// Payload contains the unpacked NAVIGATOR_SEARCH_SAVE fields.
type Payload struct {
	// Code identifies the search context.
	Code string
	// Data stores the saved query.
	Data string
}

// Definition describes the NAVIGATOR_SEARCH_SAVE payload fields.
var Definition = codec.Definition{
	codec.Named("code", codec.StringField),
	codec.Named("data", codec.StringField),
}

// Decode unpacks a NAVIGATOR_SEARCH_SAVE packet payload.
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
