// Package buyresult contains the MARKETPLACE_AFTER_ORDER_STATUS outbound packet.
package buyresult

import "github.com/niflaot/pixels/networking/codec"

// Header identifies MARKETPLACE_AFTER_ORDER_STATUS.
const Header uint16 = 2032

// Encode creates MARKETPLACE_AFTER_ORDER_STATUS.
func Encode(result int32, newOfferID int64, newPrice int64, requestedOfferID int64) (codec.Packet, error) {
	payload, err := codec.AppendPayload(nil, codec.Definition{codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field}, codec.Int32(result), codec.Int32(int32(newOfferID)), codec.Int32(int32(newPrice)), codec.Int32(int32(requestedOfferID)))
	return codec.Packet{Header: Header, Payload: payload}, err
}
