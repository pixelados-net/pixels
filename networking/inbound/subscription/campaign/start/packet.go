// Package start decodes the retired START_CAMPAIGN request.
package start

import (
	"github.com/niflaot/pixels/networking/codec"
	"github.com/niflaot/pixels/networking/inbound"
)

// Header identifies START_CAMPAIGN.
const Header uint16 = 1697

// Definition describes the abandoned campaign code field.
var Definition = codec.Definition{codec.Named("campaignCode", codec.StringField)}

// Decode validates the packet and returns its compatibility-only campaign code.
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
