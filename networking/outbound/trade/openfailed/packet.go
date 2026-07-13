// Package openfailed contains the TRADE_OPEN_FAILED outbound packet.
package openfailed

import "github.com/niflaot/pixels/networking/codec"

// Header identifies TRADE_OPEN_FAILED.
const Header uint16 = 217

// Encode creates TRADE_OPEN_FAILED.
func Encode(reason int32, otherUsername string) (codec.Packet, error) {
	payload, err := codec.AppendPayload(nil, codec.Definition{codec.Int32Field, codec.StringField}, codec.Int32(reason), codec.String(otherUsername))
	return codec.Packet{Header: Header, Payload: payload}, err
}
