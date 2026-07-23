// Package posted contains the MARKETPLACE_ITEM_POSTED outbound packet.
package posted

import "github.com/niflaot/pixels/networking/codec"

// Header identifies MARKETPLACE_ITEM_POSTED.
const Header uint16 = 1359

// Encode creates MARKETPLACE_ITEM_POSTED.
func Encode(result int32) (codec.Packet, error) {
	payload, err := codec.AppendPayload(nil, codec.Definition{codec.Int32Field}, codec.Int32(result))
	return codec.Packet{Header: Header, Payload: payload}, err
}
