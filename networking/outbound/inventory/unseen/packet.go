// Package unseen contains the UNSEEN_ITEMS outbound packet.
package unseen

import (
	"errors"
	"math"

	"github.com/niflaot/pixels/networking/codec"
)

const (
	// Header is the UNSEEN_ITEMS packet identifier.
	Header uint16 = 2103

	// ownedFurnitureCategory identifies newly owned furniture.
	ownedFurnitureCategory int32 = 1
)

var (
	// ErrItemIDRange reports an owned item id that cannot fit the wire integer.
	ErrItemIDRange = errors.New("unseen item id exceeds protocol range")
)

// EncodeOwned creates an UNSEEN_ITEMS packet for newly owned furniture ids.
func EncodeOwned(itemIDs []int64) (codec.Packet, error) {
	for _, itemID := range itemIDs {
		if itemID < math.MinInt32 || itemID > math.MaxInt32 {
			return codec.Packet{}, ErrItemIDRange
		}
	}
	payload, err := codec.AppendPayload(nil, codec.Definition{
		codec.Named("categoryCount", codec.Int32Field),
		codec.Named("category", codec.Int32Field),
		codec.Named("itemCount", codec.Int32Field),
	}, codec.Int32(1), codec.Int32(ownedFurnitureCategory), codec.Int32(int32(len(itemIDs))))
	if err != nil {
		return codec.Packet{}, err
	}
	for _, itemID := range itemIDs {
		payload, err = codec.AppendPayload(payload, codec.Definition{codec.Named("itemId", codec.Int32Field)}, codec.Int32(int32(itemID)))
		if err != nil {
			return codec.Packet{}, err
		}
	}

	return codec.Packet{Header: Header, Payload: payload}, nil
}
