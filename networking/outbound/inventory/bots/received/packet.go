// Package received encodes BOT_RECEIVED.
package received

import (
	"github.com/niflaot/pixels/networking/codec"
	outadd "github.com/niflaot/pixels/networking/outbound/inventory/bots/add"
	outlist "github.com/niflaot/pixels/networking/outbound/inventory/bots/list"
)

// Header identifies BOT_RECEIVED.
const Header uint16 = 3684

// Encode creates BOT_RECEIVED.
func Encode(bot outlist.Bot, boughtAsGift bool) (codec.Packet, error) {
	added, err := outadd.Encode(bot, false)
	if err != nil {
		return codec.Packet{}, err
	}
	prefix, err := codec.AppendPayload(nil, codec.Definition{codec.BooleanField}, codec.Bool(boughtAsGift))
	if err != nil {
		return codec.Packet{}, err
	}
	return codec.Packet{Header: Header, Payload: append(prefix, added.Payload[:len(added.Payload)-1]...)}, nil
}
