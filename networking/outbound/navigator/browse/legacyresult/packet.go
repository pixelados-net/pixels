// Package legacyresult contains the GUEST_ROOM_SEARCH_RESULT outbound packet.
package legacyresult

import (
	"github.com/niflaot/pixels/networking/codec"
	"github.com/niflaot/pixels/networking/outbound/navigator/browse/roomcard"
)

// Header identifies GUEST_ROOM_SEARCH_RESULT.
const Header uint16 = 52

// Definition describes the legacy result prefix.
var Definition = codec.Definition{
	codec.Named("searchType", codec.Int32Field),
	codec.Named("searchParam", codec.StringField),
	codec.Named("roomCount", codec.Int32Field),
}

// trailerDefinition describes the absent room advertisement marker.
var trailerDefinition = codec.Definition{codec.Named("hasAd", codec.BooleanField)}

// Encode creates a legacy room search result without an advertisement.
func Encode(searchType int32, searchParam string, rooms []roomcard.Card) (codec.Packet, error) {
	payload, err := codec.AppendPayload(nil, Definition, codec.Int32(searchType), codec.String(searchParam), codec.Int32(int32(len(rooms))))
	if err != nil {
		return codec.Packet{}, err
	}
	for _, room := range rooms {
		payload, err = roomcard.Append(payload, room)
		if err != nil {
			return codec.Packet{}, err
		}
	}
	payload, err = codec.AppendPayload(payload, trailerDefinition, codec.Bool(false))
	if err != nil {
		return codec.Packet{}, err
	}
	return codec.Packet{Header: Header, Payload: payload}, nil
}
