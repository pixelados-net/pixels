// Package add contains the USER_FURNITURE_ADD outbound packet.
package add

import (
	"github.com/niflaot/pixels/networking/codec"
	outlist "github.com/niflaot/pixels/networking/outbound/inventory/furniture/list"
)

// Header identifies USER_FURNITURE_ADD.
const Header uint16 = 104

// Encode creates an incremental furniture inventory add or update packet.
func Encode(item outlist.Item) (codec.Packet, error) {
	payload, err := outlist.AppendItem(nil, item)
	if err != nil {
		return codec.Packet{}, err
	}

	return codec.Packet{Header: Header, Payload: payload}, nil
}
