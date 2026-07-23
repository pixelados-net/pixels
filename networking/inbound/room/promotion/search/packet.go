// Package search decodes ROOM_AD_SEARCH requests.
package search

import (
	"github.com/niflaot/pixels/networking/codec"
	"github.com/niflaot/pixels/networking/inbound"
)

// Header identifies ROOM_AD_SEARCH.
const Header uint16 = 2809

// Payload contains the renderer's two search filters.
type Payload struct {
	CategoryID int32
	Offset     int32
}

// Definition describes the two integer filters.
var Definition = codec.Definition{codec.Named("categoryId", codec.Int32Field), codec.Named("offset", codec.Int32Field)}

// Decode returns one room-ad search request.
func Decode(packet codec.Packet) (Payload, error) {
	if err := inbound.ValidateHeader(packet, Header); err != nil {
		return Payload{}, err
	}
	v, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return Payload{}, err
	}
	return Payload{CategoryID: v[0].Int32, Offset: v[1].Int32}, nil
}
