// Package offers contains the MARKETPLACE_OFFERS outbound packet.
package offers

import (
	marketcore "github.com/niflaot/pixels/internal/realm/marketplace/core"
	"github.com/niflaot/pixels/networking/codec"
	marketwire "github.com/niflaot/pixels/networking/outbound/marketplace/record"
)

// Header identifies MARKETPLACE_OFFERS.
const Header uint16 = 680

// Encode creates MARKETPLACE_OFFERS.
func Encode(values []marketcore.Offer, total int32) (codec.Packet, error) {
	payload, err := codec.AppendPayload(nil, codec.Definition{codec.Int32Field}, codec.Int32(int32(len(values))))
	if err != nil {
		return codec.Packet{}, err
	}
	for _, offer := range values {
		payload, err = marketwire.AppendOffer(payload, offer, false)
		if err != nil {
			return codec.Packet{}, err
		}
	}
	payload, err = codec.AppendPayload(payload, codec.Definition{codec.Int32Field}, codec.Int32(total))
	return codec.Packet{Header: Header, Payload: payload}, err
}
