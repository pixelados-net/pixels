// Package request contains the GET_CATALOG_PAGE inbound packet.
package request

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the GET_CATALOG_PAGE packet identifier.
	Header uint16 = 412
)

// Payload contains the unpacked GET_CATALOG_PAGE fields.
type Payload struct {
	// PageID identifies the requested catalog page.
	PageID int32
	// OfferID identifies an optional highlighted offer.
	OfferID int32
	// Mode identifies the requested catalog mode.
	Mode string
}

// Definition describes the GET_CATALOG_PAGE payload fields.
var Definition = codec.Definition{
	codec.Named("pageId", codec.Int32Field),
	codec.Named("offerId", codec.Int32Field),
	codec.Named("mode", codec.StringField),
}

// Decode unpacks a GET_CATALOG_PAGE packet payload.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return Payload{}, err
	}

	return Payload{PageID: values[0].Int32, OfferID: values[1].Int32, Mode: values[2].String}, nil
}
