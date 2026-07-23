// Package info decodes GET_ROOM_AD_PURCHASE_INFO requests.
package info

import (
	"github.com/niflaot/pixels/networking/codec"
	"github.com/niflaot/pixels/networking/inbound"
)

// Header identifies GET_ROOM_AD_PURCHASE_INFO.
const Header uint16 = 1075

// Definition describes the header-only request.
var Definition = codec.Definition{}

// Decode validates one purchase-info request.
func Decode(packet codec.Packet) error {
	if err := inbound.ValidateHeader(packet, Header); err != nil {
		return err
	}
	_, err := codec.DecodePacketExact(packet, Definition)
	return err
}
