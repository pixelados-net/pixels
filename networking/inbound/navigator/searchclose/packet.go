// Package searchclose contains the NAVIGATOR_SEARCH_CLOSE inbound packet.
package searchclose

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the NAVIGATOR_SEARCH_CLOSE packet identifier.
	Header uint16 = 1834
)

// Payload contains the unpacked NAVIGATOR_SEARCH_CLOSE fields.
type Payload struct {
	// Code identifies the category to close.
	Code string
}

// Definition describes the NAVIGATOR_SEARCH_CLOSE payload fields.
var Definition = codec.Definition{codec.Named("code", codec.StringField)}

// Decode unpacks a NAVIGATOR_SEARCH_CLOSE packet payload.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return Payload{}, err
	}
	return Payload{Code: values[0].String}, nil
}
