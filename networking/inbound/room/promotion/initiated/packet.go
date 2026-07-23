// Package initiated decodes ROOM_AD_PURCHASE_INITIATED requests.
package initiated

import (
	"github.com/niflaot/pixels/networking/codec"
	"github.com/niflaot/pixels/networking/inbound"
)

// Header identifies ROOM_AD_PURCHASE_INITIATED.
const Header uint16 = 2283

// Definition describes the header-only telemetry request.
var Definition = codec.Definition{}

// Decode validates one purchase-initiation telemetry packet.
func Decode(packet codec.Packet) error {
	if err := inbound.ValidateHeader(packet, Header); err != nil {
		return err
	}
	_, err := codec.DecodePacketExact(packet, Definition)
	return err
}
