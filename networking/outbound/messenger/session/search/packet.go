// Package search contains MESSENGER_SEARCH.
package search

import "github.com/niflaot/pixels/networking/codec"

// Header identifies MESSENGER_SEARCH.
const Header uint16 = 973

// Result contains one Nitro messenger search result.
type Result struct {
	// PlayerID identifies the matching player.
	PlayerID int64
	// Username stores the visible player name.
	Username string
	// Motto stores the visible motto.
	Motto string
	// Online reports whether the player is connected.
	Online bool
	// CanFollow reports whether the player can be followed.
	CanFollow bool
	// LastOnline stores the optional last-online text.
	LastOnline string
	// Gender stores Nitro's numeric gender value.
	Gender int32
	// Look stores the avatar figure string.
	Look string
	// RealName stores the optional real-name field.
	RealName string
}

// Encode creates MESSENGER_SEARCH with friends separated from other matches.
func Encode(friends []Result, others []Result) (codec.Packet, error) {
	payload, err := appendResults(nil, friends)
	if err != nil {
		return codec.Packet{}, err
	}
	payload, err = appendResults(payload, others)
	if err != nil {
		return codec.Packet{}, err
	}
	return codec.Packet{Header: Header, Payload: payload}, nil
}

// appendResults appends one count-prefixed search-result collection.
func appendResults(dst []byte, results []Result) ([]byte, error) {
	dst, err := codec.AppendPayload(dst, codec.Definition{codec.Int32Field}, codec.Int32(int32(len(results))))
	if err != nil {
		return dst, err
	}
	for _, result := range results {
		dst, err = codec.AppendPayload(dst, resultDefinition,
			codec.Int32(int32(result.PlayerID)), codec.String(result.Username), codec.String(result.Motto),
			codec.Bool(result.Online), codec.Bool(result.CanFollow), codec.String(result.LastOnline),
			codec.Int32(result.Gender), codec.String(result.Look), codec.String(result.RealName),
		)
		if err != nil {
			return dst, err
		}
	}
	return dst, nil
}

// resultDefinition describes one Nitro search result.
var resultDefinition = codec.Definition{
	codec.Int32Field, codec.StringField, codec.StringField, codec.BooleanField,
	codec.BooleanField, codec.StringField, codec.Int32Field, codec.StringField,
	codec.StringField,
}
