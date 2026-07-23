// Package searchresult contains the NAVIGATOR_SEARCH outbound packet.
package searchresult

import (
	"github.com/niflaot/pixels/networking/codec"
	"github.com/niflaot/pixels/networking/outbound/navigator/browse/roomcard"
)

const (
	// Header is the NAVIGATOR_SEARCH packet identifier.
	Header uint16 = 2690
)

// ResultList contains one navigator result list group.
type ResultList struct {
	// Code identifies the result list.
	Code string
	// Data stores the result list label or query.
	Data string
	// Action stores the list action.
	Action int32
	// Closed reports whether the list starts collapsed.
	Closed bool
	// Mode stores the list display mode.
	Mode int32
	// Rooms stores the room cards in this list.
	Rooms []roomcard.Card
}

// Definition describes the NAVIGATOR_SEARCH payload fields.
var Definition = codec.Definition{
	codec.Named("code", codec.StringField),
	codec.Named("data", codec.StringField),
	codec.Named("resultListCount", codec.Int32Field),
}

// ResultListDefinition describes one result list header.
var ResultListDefinition = codec.Definition{
	codec.Named("code", codec.StringField),
	codec.Named("data", codec.StringField),
	codec.Named("action", codec.Int32Field),
	codec.Named("closed", codec.BooleanField),
	codec.Named("mode", codec.Int32Field),
	codec.Named("roomCount", codec.Int32Field),
}

// Encode creates a NAVIGATOR_SEARCH packet.
func Encode(code string, data string, lists []ResultList) (codec.Packet, error) {
	payload, err := codec.AppendPayload(nil, Definition,
		codec.String(code),
		codec.String(data),
		codec.Int32(int32(len(lists))),
	)
	if err != nil {
		return codec.Packet{}, err
	}
	for _, list := range lists {
		payload, err = appendList(payload, list)
		if err != nil {
			return codec.Packet{}, err
		}
	}

	return codec.Packet{Header: Header, Payload: payload}, nil
}

// appendList appends one result list group.
func appendList(dst []byte, list ResultList) ([]byte, error) {
	dst, err := codec.AppendPayload(dst, ResultListDefinition,
		codec.String(list.Code),
		codec.String(list.Data),
		codec.Int32(list.Action),
		codec.Bool(list.Closed),
		codec.Int32(list.Mode),
		codec.Int32(int32(len(list.Rooms))),
	)
	if err != nil {
		return dst, err
	}
	for _, room := range list.Rooms {
		dst, err = roomcard.Append(dst, room)
		if err != nil {
			return dst, err
		}
	}

	return dst, nil
}
