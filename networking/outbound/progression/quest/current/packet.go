// Package current encodes QUEST_CURRENT responses.
package current

import (
	"github.com/niflaot/pixels/networking/codec"
	questdata "github.com/niflaot/pixels/networking/outbound/progression/quest/data"
)

// Header identifies QUEST_CURRENT.
const Header uint16 = 230

// Encode creates one current quest response.
func Encode(quest questdata.Quest) (codec.Packet, error) {
	payload, err := questdata.Append(nil, quest)
	return codec.Packet{Header: Header, Payload: payload}, err
}
