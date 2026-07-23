// Package tags contains the ROOM_POPULAR_TAGS_RESULT outbound packet.
package tags

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the ROOM_POPULAR_TAGS_RESULT packet identifier.
	Header uint16 = 2012
)

// Entry contains one room tag result.
type Entry struct {
	// Tag stores the tag text.
	Tag string
	// Count stores the tag usage count.
	Count int32
}

// Definition describes the ROOM_POPULAR_TAGS_RESULT payload fields.
var Definition = codec.Definition{codec.Named("entryCount", codec.Int32Field)}

// EntryDefinition describes one ROOM_POPULAR_TAGS_RESULT entry.
var EntryDefinition = codec.Definition{
	codec.Named("tag", codec.StringField),
	codec.Named("count", codec.Int32Field),
}

// Encode creates a ROOM_POPULAR_TAGS_RESULT packet.
func Encode(entries []Entry) (codec.Packet, error) {
	payload, err := codec.AppendPayload(nil, Definition, codec.Int32(int32(len(entries))))
	if err != nil {
		return codec.Packet{}, err
	}
	for _, entry := range entries {
		payload, err = codec.AppendPayload(payload, EntryDefinition, codec.String(entry.Tag), codec.Int32(entry.Count))
		if err != nil {
			return codec.Packet{}, err
		}
	}

	return codec.Packet{Header: Header, Payload: payload}, nil
}
