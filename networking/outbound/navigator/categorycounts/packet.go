// Package categorycounts contains the CATEGORIES_WITH_VISITOR_COUNT outbound packet.
package categorycounts

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the CATEGORIES_WITH_VISITOR_COUNT packet identifier.
	Header uint16 = 1455
)

// Entry contains one category visitor count.
type Entry struct {
	// CategoryID identifies the category.
	CategoryID int32
	// CurrentVisitorCount stores current visitors.
	CurrentVisitorCount int32
	// MaxVisitorCount stores the maximum visitor count.
	MaxVisitorCount int32
}

// Definition describes the CATEGORIES_WITH_VISITOR_COUNT payload fields.
var Definition = codec.Definition{codec.Named("entryCount", codec.Int32Field)}

// EntryDefinition describes one category count entry.
var EntryDefinition = codec.Definition{
	codec.Named("categoryId", codec.Int32Field),
	codec.Named("currentVisitorCount", codec.Int32Field),
	codec.Named("maxVisitorCount", codec.Int32Field),
}

// Encode creates a CATEGORIES_WITH_VISITOR_COUNT packet.
func Encode(entries []Entry) (codec.Packet, error) {
	payload, err := codec.AppendPayload(nil, Definition, codec.Int32(int32(len(entries))))
	if err != nil {
		return codec.Packet{}, err
	}
	for _, entry := range entries {
		payload, err = codec.AppendPayload(payload, EntryDefinition,
			codec.Int32(entry.CategoryID),
			codec.Int32(entry.CurrentVisitorCount),
			codec.Int32(entry.MaxVisitorCount),
		)
		if err != nil {
			return codec.Packet{}, err
		}
	}

	return codec.Packet{Header: Header, Payload: payload}, nil
}
