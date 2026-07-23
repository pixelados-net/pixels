// Package countdown decodes GET_SECONDS_UNTIL requests.
package countdown

import (
	"github.com/niflaot/pixels/networking/codec"
	"github.com/niflaot/pixels/networking/inbound"
)

// Header identifies GET_SECONDS_UNTIL.
const Header uint16 = 271

// Definition describes the target date string.
var Definition = codec.Definition{codec.Named("targetDate", codec.StringField)}

// Decode returns the requested target date.
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
