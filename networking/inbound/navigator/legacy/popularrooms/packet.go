// Package popularrooms decodes POPULAR_ROOMS_SEARCH requests.
package popularrooms

import (
	"github.com/niflaot/pixels/networking/codec"
	"github.com/niflaot/pixels/networking/inbound"
)

// Header identifies POPULAR_ROOMS_SEARCH.
const Header uint16 = 2758

// Request stores the legacy popular-room filters.
type Request struct {
	// Query stores the room filter string.
	Query string
	// PageIndex stores the requested result page.
	PageIndex int32
}

// Definition describes the legacy popular-room filters.
var Definition = codec.Definition{codec.Named("query", codec.StringField), codec.Named("pageIndex", codec.Int32Field)}

// Decode returns the requested popular-room filters.
func Decode(packet codec.Packet) (Request, error) {
	if err := inbound.ValidateHeader(packet, Header); err != nil {
		return Request{}, err
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return Request{}, err
	}
	return Request{Query: values[0].String, PageIndex: values[1].Int32}, nil
}
