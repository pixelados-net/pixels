// Package stats contains the MARKETPLACE_ITEM_STATS outbound packet.
package stats

import (
	marketcore "github.com/niflaot/pixels/internal/realm/marketplace/core"
	"github.com/niflaot/pixels/networking/codec"
	"time"
)

// Header identifies MARKETPLACE_ITEM_STATS.
const Header uint16 = 725

// Encode creates MARKETPLACE_ITEM_STATS.
func Encode(value marketcore.Stats, category int32, itemID int32) (codec.Packet, error) {
	payload, err := codec.AppendPayload(nil, codec.Definition{codec.Int32Field, codec.Int32Field, codec.Int32Field}, codec.Int32(int32(value.AveragePrice)), codec.Int32(value.OpenCount), codec.Int32(int32(len(value.History))))
	if err != nil {
		return codec.Packet{}, err
	}
	today := time.Now().UTC()
	for _, day := range value.History {
		offset := int32(today.Sub(day.Day).Hours() / 24)
		payload, err = codec.AppendPayload(payload, codec.Definition{codec.Int32Field, codec.Int32Field, codec.Int32Field}, codec.Int32(offset), codec.Int32(int32(day.AverageRawPrice)), codec.Int32(day.SoldCount))
		if err != nil {
			return codec.Packet{}, err
		}
	}
	payload, err = codec.AppendPayload(payload, codec.Definition{codec.Int32Field, codec.Int32Field}, codec.Int32(category), codec.Int32(itemID))
	return codec.Packet{Header: Header, Payload: payload}, err
}
