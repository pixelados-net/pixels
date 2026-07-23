// Package outside decodes REQUESTFURNIINVENTORYWHENNOTINROOM requests.
package outside

import (
	"github.com/niflaot/pixels/networking/codec"
	"github.com/niflaot/pixels/networking/inbound"
)

// Header identifies REQUESTFURNIINVENTORYWHENNOTINROOM.
const Header uint16 = 3500

// Definition describes the header-only inventory alias.
var Definition = codec.Definition{}

// Decode validates one outside-room furniture inventory request.
func Decode(packet codec.Packet) error {
	if err := inbound.ValidateHeader(packet, Header); err != nil {
		return err
	}
	_, err := codec.DecodePacketExact(packet, Definition)
	return err
}
