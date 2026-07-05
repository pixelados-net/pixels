// Package eventcategories contains the NAVIGATOR_EVENT_CATEGORIES outbound packet.
package eventcategories

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the NAVIGATOR_EVENT_CATEGORIES packet identifier.
	Header uint16 = 3244
)

// Category contains one navigator event category.
type Category struct {
	// ID identifies the event category.
	ID int32
	// Name stores the event category name.
	Name string
	// Visible reports whether the category is visible.
	Visible bool
}

// Definition describes the NAVIGATOR_EVENT_CATEGORIES payload fields.
var Definition = codec.Definition{codec.Named("categoryCount", codec.Int32Field)}

// CategoryDefinition describes one event category entry.
var CategoryDefinition = codec.Definition{
	codec.Named("id", codec.Int32Field),
	codec.Named("name", codec.StringField),
	codec.Named("visible", codec.BooleanField),
}

// Encode creates a NAVIGATOR_EVENT_CATEGORIES packet.
func Encode(categories []Category) (codec.Packet, error) {
	payload, err := codec.AppendPayload(nil, Definition, codec.Int32(int32(len(categories))))
	if err != nil {
		return codec.Packet{}, err
	}
	for _, category := range categories {
		payload, err = codec.AppendPayload(payload, CategoryDefinition,
			codec.Int32(category.ID),
			codec.String(category.Name),
			codec.Bool(category.Visible),
		)
		if err != nil {
			return codec.Packet{}, err
		}
	}

	return codec.Packet{Header: Header, Payload: payload}, nil
}
