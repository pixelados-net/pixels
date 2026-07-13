// Package tokens contains the MARKETPLACE_TOKENS outbound packet.
package tokens

import "github.com/niflaot/pixels/networking/codec"

// Header identifies MARKETPLACE_TOKENS.
const Header uint16 = 54

// Encode creates MARKETPLACE_TOKENS.
func Encode(result int32, count int32) (codec.Packet, error) {
	payload, err := codec.AppendPayload(nil, codec.Definition{codec.Int32Field, codec.Int32Field}, codec.Int32(result), codec.Int32(count))
	return codec.Packet{Header: Header, Payload: payload}, err
}
