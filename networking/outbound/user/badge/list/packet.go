// Package list encodes player badge inventory snapshots.
package list

import "github.com/niflaot/pixels/networking/codec"

// Header identifies USER_BADGES responses.
const Header uint16 = 717

// Badge describes one owned badge and its optional active slot.
type Badge struct {
	// ID identifies the durable badge inventory entry.
	ID int32
	// Code identifies the badge image asset.
	Code string
	// Slot stores the active position or zero when unequipped.
	Slot int32
}

// Encode creates a complete badge inventory response.
func Encode(badges []Badge) (codec.Packet, error) {
	payload, err := codec.AppendPayload(nil, codec.Definition{codec.Int32Field}, codec.Int32(int32(len(badges))))
	for _, badge := range badges {
		if err != nil {
			return codec.Packet{}, err
		}
		payload, err = codec.AppendPayload(payload, codec.Definition{codec.Int32Field, codec.StringField}, codec.Int32(badge.ID), codec.String(badge.Code))
	}
	active := int32(0)
	for _, badge := range badges {
		if badge.Slot > 0 {
			active++
		}
	}
	if err == nil {
		payload, err = codec.AppendPayload(payload, codec.Definition{codec.Int32Field}, codec.Int32(active))
	}
	for _, badge := range badges {
		if err != nil {
			return codec.Packet{}, err
		}
		if badge.Slot > 0 {
			payload, err = codec.AppendPayload(payload, codec.Definition{codec.Int32Field, codec.StringField}, codec.Int32(badge.Slot), codec.String(badge.Code))
		}
	}
	return codec.Packet{Header: Header, Payload: payload}, err
}
