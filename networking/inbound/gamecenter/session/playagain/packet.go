// Package playagain decodes GAME2PLAYAGAINMESSAGE requests.
package playagain

import (
	"github.com/niflaot/pixels/networking/codec"
	"github.com/niflaot/pixels/networking/inbound"
)

// Header identifies GAME2PLAYAGAINMESSAGE.
const Header uint16 = 3196

// Decode validates one header-only play again request.
func Decode(packet codec.Packet) error {
	if err := inbound.ValidateHeader(packet, Header); err != nil {
		return err
	}
	_, err := codec.DecodePacketExact(packet, codec.Definition{})
	return err
}
