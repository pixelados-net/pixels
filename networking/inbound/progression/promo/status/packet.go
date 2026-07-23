// Package status decodes GETISBADGEREQUESTFULFILLED requests.
package status

import (
	"github.com/niflaot/pixels/networking/codec"
	"github.com/niflaot/pixels/networking/inbound"
)

// Header identifies GETISBADGEREQUESTFULFILLED.
const Header uint16 = 1364

// Decode returns the requested promotion code.
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
