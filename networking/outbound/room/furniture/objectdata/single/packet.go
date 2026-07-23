// Package single encodes the FURNITURE_DATA outbound packet.
package single

import (
	"strconv"

	"github.com/niflaot/pixels/networking/codec"
	"github.com/niflaot/pixels/networking/outbound/furniture/stuffdata"
)

// Header is the FURNITURE_DATA packet identifier.
const Header uint16 = 2547

// Encode creates one typed furniture-data update.
func Encode(itemID int64, data *stuffdata.Data) (codec.Packet, error) {
	if data == nil {
		return codec.Packet{}, codec.ErrInvalidField
	}
	payload, err := codec.AppendPayload(nil, codec.Definition{codec.StringField}, codec.String(strconv.FormatInt(itemID, 10)))
	if err != nil {
		return codec.Packet{}, err
	}
	payload, err = data.Append(payload)
	return codec.Packet{Header: Header, Payload: payload}, err
}
