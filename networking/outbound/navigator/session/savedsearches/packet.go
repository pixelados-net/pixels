// Package savedsearches contains the NAVIGATOR_SEARCHES outbound packet.
package savedsearches

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the NAVIGATOR_SEARCHES packet identifier.
	Header uint16 = 3984
)

// Search contains one saved navigator search.
type Search struct {
	// ID identifies the saved search.
	ID int32
	// Code stores the search code.
	Code string
	// Filter stores the search filter.
	Filter string
	// Localization stores the display localization.
	Localization string
}

// Definition describes the NAVIGATOR_SEARCHES payload fields.
var Definition = codec.Definition{codec.Named("searchCount", codec.Int32Field)}

// SearchDefinition describes one saved search entry.
var SearchDefinition = codec.Definition{
	codec.Named("id", codec.Int32Field),
	codec.Named("code", codec.StringField),
	codec.Named("filter", codec.StringField),
	codec.Named("localization", codec.StringField),
}

// Encode creates a NAVIGATOR_SEARCHES packet.
func Encode(searches []Search) (codec.Packet, error) {
	payload, err := codec.AppendPayload(nil, Definition, codec.Int32(int32(len(searches))))
	if err != nil {
		return codec.Packet{}, err
	}
	for _, search := range searches {
		payload, err = codec.AppendPayload(payload, SearchDefinition,
			codec.Int32(search.ID),
			codec.String(search.Code),
			codec.String(search.Filter),
			codec.String(search.Localization),
		)
		if err != nil {
			return codec.Packet{}, err
		}
	}

	return codec.Packet{Header: Header, Payload: payload}, nil
}
