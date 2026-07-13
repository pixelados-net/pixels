// Package closed contains the TRADE_CLOSED outbound packet.
package closed

import "github.com/niflaot/pixels/networking/codec"

// Header identifies TRADE_CLOSED.
const Header uint16 = 1373

// Encode creates TRADE_CLOSED.
func Encode(playerID int64, reason int32) (codec.Packet, error) {
	payload, err := codec.AppendPayload(nil, codec.Definition{codec.Int32Field, codec.Int32Field}, codec.Int32(int32(playerID)), codec.Int32(reason))
	return codec.Packet{Header: Header, Payload: payload}, err
}
