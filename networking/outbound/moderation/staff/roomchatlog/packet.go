// Package roomchatlog contains room moderator chat history.
package roomchatlog

import (
	"github.com/niflaot/pixels/networking/codec"
	"github.com/niflaot/pixels/networking/outbound/moderation/chatrecord"
)

// Header identifies MODTOOL_ROOM_CHATLOG.
const Header uint16 = 3434

// Encode creates one room chat record packet.
func Encode(record chatrecord.Record) (codec.Packet, error) {
	payload, err := chatrecord.Append(nil, record)
	return codec.Packet{Header: Header, Payload: payload}, err
}
