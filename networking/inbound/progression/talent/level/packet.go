// Package level decodes TALENT_TRACK_GET_LEVEL requests.
package level

import (
	"github.com/niflaot/pixels/networking/codec"
	"github.com/niflaot/pixels/networking/inbound"
)

// Header identifies TALENT_TRACK_GET_LEVEL.
const Header uint16 = 2127

// Decode returns the requested talent track name.
func Decode(packet codec.Packet) (string, error) {
	if err := inbound.ValidateHeader(packet, Header); err != nil {
		return "", err
	}
	values, err := codec.DecodePacketExact(packet, codec.Definition{codec.StringField})
	if err != nil {
		return "", err
	}
	return values[0].String, nil
}
