// Package batch encodes the OBJECTS_DATA_UPDATE outbound packet.
package batch

import (
	"github.com/niflaot/pixels/networking/codec"
	"github.com/niflaot/pixels/networking/outbound/furniture/stuffdata"
)

// Header is the OBJECTS_DATA_UPDATE packet identifier.
const Header uint16 = 1453

// Encode creates a batch typed furniture-data update.
func Encode(itemIDs []int64, data []*stuffdata.Data) (codec.Packet, error) {
	if len(itemIDs) != len(data) {
		return codec.Packet{}, codec.ErrInvalidField
	}
	payload, err := codec.AppendPayload(nil, codec.Definition{codec.Int32Field}, codec.Int32(int32(len(itemIDs))))
	if err != nil {
		return codec.Packet{}, err
	}
	for index, itemID := range itemIDs {
		if data[index] == nil {
			return codec.Packet{}, codec.ErrInvalidField
		}
		payload, err = codec.AppendPayload(payload, codec.Definition{codec.Int32Field}, codec.Int32(int32(itemID)))
		if err != nil {
			return codec.Packet{}, err
		}
		payload, err = data[index].Append(payload)
		if err != nil {
			return codec.Packet{}, err
		}
	}
	return codec.Packet{Header: Header, Payload: payload}, nil
}
