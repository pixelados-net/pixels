// Package roomtextsearch decodes ROOM_TEXT_SEARCH requests.
package roomtextsearch

import (
	"github.com/niflaot/pixels/networking/codec"
	"github.com/niflaot/pixels/networking/inbound"
)

// Header identifies ROOM_TEXT_SEARCH.
const Header uint16 = 3943

// Definition describes the legacy room text query.
var Definition = codec.Definition{codec.Named("query", codec.StringField)}

// Decode returns the requested room text query.
func Decode(packet codec.Packet) (string, error) {
	if err := inbound.ValidateHeader(packet, Header); err != nil {
		return "", err
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return "", err
	}
	return values[0].String, nil
}
