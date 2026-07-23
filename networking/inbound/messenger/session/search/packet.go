// Package search contains HABBO_SEARCH.
package search

import "github.com/niflaot/pixels/networking/codec"

// Header identifies HABBO_SEARCH.
const Header uint16 = 1210

// Decode unpacks the search term.
func Decode(packet codec.Packet) (string, error) {
	if packet.Header != Header {
		return "", codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, codec.Definition{codec.Named("term", codec.StringField)})
	if err != nil {
		return "", err
	}
	return values[0].String, nil
}
