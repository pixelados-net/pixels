// Package add encodes ADD_BOT_TO_INVENTORY.
package add

import (
	"github.com/niflaot/pixels/networking/codec"
	outlist "github.com/niflaot/pixels/networking/outbound/inventory/bots/list"
)

// Header identifies ADD_BOT_TO_INVENTORY.
const Header uint16 = 1352

// Encode creates ADD_BOT_TO_INVENTORY.
func Encode(bot outlist.Bot, openInventory bool) (codec.Packet, error) {
	packet, err := outlist.Encode([]outlist.Bot{bot})
	if err != nil {
		return codec.Packet{}, err
	}
	payload := packet.Payload[4:]
	payload, err = codec.AppendPayload(payload, codec.Definition{codec.BooleanField}, codec.Bool(openInventory))
	return codec.Packet{Header: Header, Payload: payload}, err
}
