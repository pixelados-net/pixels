// Package page contains the CATALOG_PAGE outbound packet.
package page

import (
	"github.com/niflaot/pixels/networking/codec"
	"github.com/niflaot/pixels/networking/outbound/catalog/offer"
)

const (
	// Header is the CATALOG_PAGE packet identifier.
	Header uint16 = 804
)

// Localization contains page layout images and text.
type Localization struct {
	// Images stores client image assets.
	Images []string
	// Texts stores localized page text.
	Texts []string
}

// Encode creates a CATALOG_PAGE packet.
func Encode(pageID int32, mode string, layout string, localization Localization, offers []offer.Offer, highlightedOfferID int32) (codec.Packet, error) {
	payload, err := codec.AppendPayload(nil, codec.Definition{
		codec.Int32Field, codec.StringField, codec.StringField, codec.Int32Field,
	}, codec.Int32(pageID), codec.String(mode), codec.String(layout), codec.Int32(int32(len(localization.Images))))
	if err != nil {
		return codec.Packet{}, err
	}
	payload, err = appendStrings(payload, localization.Images)
	if err != nil {
		return codec.Packet{}, err
	}
	payload, err = codec.AppendPayload(payload, codec.Definition{codec.Int32Field}, codec.Int32(int32(len(localization.Texts))))
	if err != nil {
		return codec.Packet{}, err
	}
	payload, err = appendStrings(payload, localization.Texts)
	if err != nil {
		return codec.Packet{}, err
	}
	payload, err = codec.AppendPayload(payload, codec.Definition{codec.Int32Field}, codec.Int32(int32(len(offers))))
	if err != nil {
		return codec.Packet{}, err
	}
	for _, catalogOffer := range offers {
		payload, err = offer.AppendPage(payload, catalogOffer)
		if err != nil {
			return codec.Packet{}, err
		}
	}
	payload, err = codec.AppendPayload(payload, codec.Definition{codec.Int32Field, codec.BooleanField}, codec.Int32(highlightedOfferID), codec.Bool(false))
	if err != nil {
		return codec.Packet{}, err
	}

	return codec.Packet{Header: Header, Payload: payload}, nil
}

// appendStrings appends protocol strings without allocating a value slice.
func appendStrings(dst []byte, values []string) ([]byte, error) {
	var err error
	for _, value := range values {
		dst, err = codec.AppendPayload(dst, codec.Definition{codec.StringField}, codec.String(value))
		if err != nil {
			return dst, err
		}
	}

	return dst, nil
}
