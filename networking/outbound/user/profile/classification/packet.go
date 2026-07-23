// Package classification contains the compatibility USER_CLASSIFICATION packet.
package classification

import (
	"github.com/niflaot/pixels/networking/codec"
)

// Header identifies USER_CLASSIFICATION.
const Header uint16 = 966

// Definition describes the classification count.
var Definition = codec.Definition{codec.Named("count", codec.Int32Field)}

// EntryDefinition describes one compatibility classification.
var EntryDefinition = codec.Definition{codec.Named("playerId", codec.Int32Field), codec.Named("username", codec.StringField), codec.Named("classType", codec.StringField)}

// Encode creates a bounded compatibility classification list.
func Encode(playerIDs []int32, usernames []string, classTypes []string) (codec.Packet, error) {
	if len(playerIDs) != len(usernames) || len(playerIDs) != len(classTypes) {
		return codec.Packet{}, codec.ErrInvalidField
	}
	payload, err := codec.AppendPayload(nil, Definition, codec.Int32(int32(len(playerIDs))))
	if err != nil {
		return codec.Packet{}, err
	}
	for index, playerID := range playerIDs {
		payload, err = codec.AppendPayload(payload, EntryDefinition, codec.Int32(playerID), codec.String(usernames[index]), codec.String(classTypes[index]))
		if err != nil {
			return codec.Packet{}, err
		}
	}
	return codec.Packet{Header: Header, Payload: payload}, nil
}
