// Package accepted contains the TRADE_ACCEPTED outbound packet.
package accepted

import "github.com/niflaot/pixels/networking/codec"

// Header identifies TRADE_ACCEPTED.
const Header uint16 = 2568

// Encode creates TRADE_ACCEPTED.
func Encode(playerID int64, value bool) (codec.Packet, error) {
	accepted := int32(0)
	if value {
		accepted = 1
	}
	payload, err := codec.AppendPayload(nil, codec.Definition{codec.Int32Field, codec.Int32Field}, codec.Int32(int32(playerID)), codec.Int32(accepted))
	return codec.Packet{Header: Header, Payload: payload}, err
}
