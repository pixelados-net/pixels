// Package state contains the SET_TARGETTED_OFFER_STATE inbound packet.
package state

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header identifies SET_TARGETTED_OFFER_STATE.
	Header uint16 = 2041
)

// Payload contains a targeted offer state mutation.
type Payload struct {
	// OfferID identifies the targeted offer.
	OfferID int32
	// State stores the requested client state.
	State int32
}

// Definition describes the packet payload.
var Definition = codec.Definition{codec.Int32Field, codec.Int32Field}

// Decode decodes a SET_TARGETTED_OFFER_STATE packet.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	v, e := codec.DecodePacketExact(packet, Definition)
	if e != nil {
		return Payload{}, e
	}
	return Payload{OfferID: v[0].Int32, State: v[1].Int32}, nil
}
