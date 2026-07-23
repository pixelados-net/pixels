// Package request decodes GET_BONUS_RARE_INFO requests.
package request

import (
	"github.com/niflaot/pixels/networking/codec"
	"github.com/niflaot/pixels/networking/inbound"
)

// Header identifies GET_BONUS_RARE_INFO.
const Header uint16 = 957

// Definition describes the header-only request.
var Definition = codec.Definition{}

// Decode validates one GET_BONUS_RARE_INFO packet.
func Decode(packet codec.Packet) error {
	if err := inbound.ValidateHeader(packet, Header); err != nil {
		return err
	}
	_, err := codec.DecodePacketExact(packet, Definition)
	return err
}
