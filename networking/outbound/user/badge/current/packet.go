// Package current encodes active badges for one room user.
package current

import "github.com/niflaot/pixels/networking/codec"

// Header identifies USER_BADGES_CURRENT responses.
const Header uint16 = 1087

// Badge describes one active badge position.
type Badge struct {
	// Slot identifies the active badge position.
	Slot int32
	// Code identifies the badge image asset.
	Code string
}

// Encode creates one room user's active badge projection.
func Encode(playerID int64, badges []Badge) (codec.Packet, error) {
	payload, err := codec.AppendPayload(nil, codec.Definition{codec.Int32Field, codec.Int32Field}, codec.Int32(int32(playerID)), codec.Int32(int32(len(badges))))
	for _, badge := range badges {
		if err != nil {
			return codec.Packet{}, err
		}
		payload, err = codec.AppendPayload(payload, codec.Definition{codec.Int32Field, codec.StringField}, codec.Int32(badge.Slot), codec.String(badge.Code))
	}
	return codec.Packet{Header: Header, Payload: payload}, err
}
