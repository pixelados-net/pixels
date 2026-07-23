// Package create contains GROUP_BUY.
package create

import (
	grouprecord "github.com/niflaot/pixels/internal/realm/group/record"
	"github.com/niflaot/pixels/networking/codec"
)

// Header identifies GROUP_BUY.
const Header uint16 = 230

const (
	// badgeValuesPerPart is Nitro's element, color, and position tuple width.
	badgeValuesPerPart = 3
	// maxBadgeParts bounds the client-controlled badge allocation.
	maxBadgeParts = 5
)

// Payload contains one complete group creator request.
type Payload struct {
	// Name stores requested identity.
	Name string
	// Description stores requested public information.
	Description string
	// RoomID identifies the home room.
	RoomID int64
	// ColorA identifies the primary editor color.
	ColorA int32
	// ColorB identifies the secondary editor color.
	ColorB int32
	// Parts stores bounded badge layers.
	Parts []grouprecord.BadgePart
}

// Decode validates scalar fields before allocating bounded badge layers.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, rest, err := codec.DecodePayload(nil, codec.Definition{codec.StringField, codec.StringField, codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field}, packet.Payload)
	if err != nil {
		return Payload{}, err
	}
	valueCount := int(values[5].Int32)
	if valueCount < badgeValuesPerPart || valueCount > maxBadgeParts*badgeValuesPerPart || valueCount%badgeValuesPerPart != 0 {
		return Payload{}, codec.ErrInvalidField
	}
	partCount := valueCount / badgeValuesPerPart
	result := Payload{Name: values[0].String, Description: values[1].String, RoomID: int64(values[2].Int32), ColorA: values[3].Int32, ColorB: values[4].Int32, Parts: make([]grouprecord.BadgePart, 0, partCount)}
	for index := range partCount {
		values, rest, err = codec.DecodePayload(values[:0], codec.Definition{codec.Int32Field, codec.Int32Field, codec.Int32Field}, rest)
		if err != nil {
			return Payload{}, err
		}
		kind := grouprecord.BadgeSymbol
		if index == 0 {
			kind = grouprecord.BadgeBase
		}
		result.Parts = append(result.Parts, grouprecord.BadgePart{Ordinal: int16(index), Kind: kind, ElementID: values[0].Int32, ColorID: values[1].Int32, Position: values[2].Int32})
	}
	if len(rest) != 0 {
		return Payload{}, codec.ErrUnexpectedPayload
	}
	return result, nil
}
