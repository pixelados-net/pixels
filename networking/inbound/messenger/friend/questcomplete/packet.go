// Package questcomplete decodes FRIEND_REQUEST_QUEST_COMPLETE requests.
package questcomplete

import (
	"github.com/niflaot/pixels/networking/codec"
	"github.com/niflaot/pixels/networking/inbound"
)

// Header identifies FRIEND_REQUEST_QUEST_COMPLETE.
const Header uint16 = 1148

// Decode validates one header-only social quest completion request.
func Decode(packet codec.Packet) error {
	if err := inbound.ValidateHeader(packet, Header); err != nil {
		return err
	}
	_, err := codec.DecodePacketExact(packet, codec.Definition{})
	return err
}
