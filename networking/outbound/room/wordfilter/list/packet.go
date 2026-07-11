// Package list contains the ROOM_GET_FILTER_WORDS outbound packet.
package list

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header identifies ROOM_GET_FILTER_WORDS.
	Header uint16 = 2937
)

// CountDefinition describes filter list metadata.
var CountDefinition = codec.Definition{codec.Named("count", codec.Int32Field)}

// WordDefinition describes one filter word.
var WordDefinition = codec.Definition{codec.Named("word", codec.StringField)}

// Encode creates a ROOM_GET_FILTER_WORDS packet.
func Encode(words []string) (codec.Packet, error) {
	payload, err := codec.AppendPayload(nil, CountDefinition, codec.Int32(int32(len(words))))
	if err != nil {
		return codec.Packet{}, err
	}
	for _, word := range words {
		payload, err = codec.AppendPayload(payload, WordDefinition, codec.String(word))
		if err != nil {
			return codec.Packet{}, err
		}
	}

	return codec.Packet{Header: Header, Payload: payload}, nil
}
