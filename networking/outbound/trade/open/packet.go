// Package open contains the TRADE_OPEN outbound packet.
package open

import "github.com/niflaot/pixels/networking/codec"

// Header identifies TRADE_OPEN.
const Header uint16 = 2505

// Encode creates TRADE_OPEN.
func Encode(firstPlayerID int64, firstCanTrade bool, secondPlayerID int64, secondCanTrade bool) (codec.Packet, error) {
	firstCapability := int32(0)
	if firstCanTrade {
		firstCapability = 1
	}
	secondCapability := int32(0)
	if secondCanTrade {
		secondCapability = 1
	}
	payload, err := codec.AppendPayload(nil, codec.Definition{codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field}, codec.Int32(int32(firstPlayerID)), codec.Int32(firstCapability), codec.Int32(int32(secondPlayerID)), codec.Int32(secondCapability))
	return codec.Packet{Header: Header, Payload: payload}, err
}
