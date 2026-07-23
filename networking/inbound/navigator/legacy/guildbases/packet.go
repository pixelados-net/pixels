// Package guildbases decodes MY_GUILD_BASES_SEARCH requests.
package guildbases

import (
	"github.com/niflaot/pixels/networking/codec"
	"github.com/niflaot/pixels/networking/inbound"
)

// Header identifies MY_GUILD_BASES_SEARCH.
const Header uint16 = 39

// Decode validates one header-only my guild bases request.
func Decode(packet codec.Packet) error {
	if err := inbound.ValidateHeader(packet, Header); err != nil {
		return err
	}
	_, err := codec.DecodePacketExact(packet, codec.Definition{})
	return err
}
