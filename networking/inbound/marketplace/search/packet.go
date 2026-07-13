// Package search contains the GET_MARKETPLACE_OFFERS inbound packet.
package search

import "github.com/niflaot/pixels/networking/codec"

// Header identifies GET_MARKETPLACE_OFFERS.
const Header uint16 = 2407

// Payload stores Marketplace filters.
type Payload struct {
	// MinimumPrice stores the inclusive buyer-price floor.
	MinimumPrice int32
	// MaximumPrice stores the inclusive buyer-price ceiling.
	MaximumPrice int32
	// Query stores the case-insensitive furniture query.
	Query string
	// SortType stores Nitro's ordering mode.
	SortType int32
}

// Decode reads Marketplace filters.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	definition := codec.Definition{codec.Named("minimumPrice", codec.Int32Field), codec.Named("maximumPrice", codec.Int32Field), codec.Named("query", codec.StringField), codec.Named("sortType", codec.Int32Field)}
	values, err := codec.DecodePacketExact(packet, definition)
	if err != nil {
		return Payload{}, err
	}
	return Payload{MinimumPrice: values[0].Int32, MaximumPrice: values[1].Int32, Query: values[2].String, SortType: values[3].Int32}, nil
}
