// Package change contains the USER_CHANGE_NAME outbound packet.
package change

import "github.com/niflaot/pixels/networking/codec"

// Header identifies USER_CHANGE_NAME.
const Header uint16 = 118

// Definition describes the fixed USER_CHANGE_NAME fields.
var Definition = codec.Definition{codec.Named("result", codec.Int32Field), codec.Named("username", codec.StringField), codec.Named("suggestionCount", codec.Int32Field)}

// SuggestionDefinition describes one username suggestion.
var SuggestionDefinition = codec.Definition{codec.Named("suggestion", codec.StringField)}

// Encode creates a USER_CHANGE_NAME packet.
func Encode(result int32, username string, suggestions []string) (codec.Packet, error) {
	payload, err := codec.AppendPayload(nil, Definition, codec.Int32(result), codec.String(username), codec.Int32(int32(len(suggestions))))
	if err != nil {
		return codec.Packet{}, err
	}
	for _, suggestion := range suggestions {
		payload, err = codec.AppendPayload(payload, SuggestionDefinition, codec.String(suggestion))
		if err != nil {
			return codec.Packet{}, err
		}
	}
	return codec.Packet{Header: Header, Payload: payload}, nil
}
