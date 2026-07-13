// Package own contains the OWN_MARKETPLACE_OFFERS outbound packet.
package own

import (
	marketcore "github.com/niflaot/pixels/internal/realm/marketplace/core"
	"github.com/niflaot/pixels/networking/codec"
	marketwire "github.com/niflaot/pixels/networking/outbound/marketplace/record"
)

// Header identifies OWN_MARKETPLACE_OFFERS.
const Header uint16 = 3884

// Encode creates OWN_MARKETPLACE_OFFERS.
func Encode(waiting int64, values []marketcore.Offer) (codec.Packet, error) {
	payload, err := codec.AppendPayload(nil, codec.Definition{codec.Int32Field, codec.Int32Field}, codec.Int32(int32(waiting)), codec.Int32(int32(len(values))))
	if err != nil {
		return codec.Packet{}, err
	}
	for _, offer := range values {
		payload, err = marketwire.AppendOffer(payload, offer, true)
		if err != nil {
			return codec.Packet{}, err
		}
	}
	return codec.Packet{Header: Header, Payload: payload}, nil
}
