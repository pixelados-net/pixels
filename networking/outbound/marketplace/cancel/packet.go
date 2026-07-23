// Package cancel contains the MARKETPLACE_CANCEL_RESULT outbound packet.
package cancel

import "github.com/niflaot/pixels/networking/codec"

// Header identifies MARKETPLACE_CANCEL_RESULT.
const Header uint16 = 3264

// Encode creates MARKETPLACE_CANCEL_RESULT.
func Encode(offerID int64, success bool) (codec.Packet, error) {
	payload, err := codec.AppendPayload(nil, codec.Definition{codec.Int32Field, codec.BooleanField}, codec.Int32(int32(offerID)), codec.Bool(success))
	return codec.Packet{Header: Header, Payload: payload}, err
}
