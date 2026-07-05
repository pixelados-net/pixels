// Package flatcats contains the NAVIGATOR_CATEGORIES outbound packet.
package flatcats

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the NAVIGATOR_CATEGORIES packet identifier.
	Header uint16 = 1562
)

// Category contains one user room category.
type Category struct {
	// ID identifies the category.
	ID int32
	// Name stores the category display key.
	Name string
	// Visible reports whether the category is visible.
	Visible bool
	// Automatic reports whether the category is automatic.
	Automatic bool
	// AutomaticCategoryKey stores the automatic category key.
	AutomaticCategoryKey string
	// GlobalCategoryKey stores the global category key.
	GlobalCategoryKey string
	// StaffOnly reports whether the category is staff-only.
	StaffOnly bool
}

// Definition describes the NAVIGATOR_CATEGORIES payload fields.
var Definition = codec.Definition{codec.Named("categoryCount", codec.Int32Field)}

// CategoryDefinition describes one room category entry.
var CategoryDefinition = codec.Definition{
	codec.Named("id", codec.Int32Field),
	codec.Named("name", codec.StringField),
	codec.Named("visible", codec.BooleanField),
	codec.Named("automatic", codec.BooleanField),
	codec.Named("automaticCategoryKey", codec.StringField),
	codec.Named("globalCategoryKey", codec.StringField),
	codec.Named("staffOnly", codec.BooleanField),
}

// Encode creates a NAVIGATOR_CATEGORIES packet.
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
			codec.Bool(category.Automatic),
			codec.String(category.AutomaticCategoryKey),
			codec.String(category.GlobalCategoryKey),
			codec.Bool(category.StaffOnly),
		)
		if err != nil {
			return codec.Packet{}, err
		}
	}

	return codec.Packet{Header: Header, Payload: payload}, nil
}
