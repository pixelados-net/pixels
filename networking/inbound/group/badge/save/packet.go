// Package save contains GROUP_SAVE_BADGE.
package save

import (
	grouprecord "github.com/niflaot/pixels/internal/realm/group/record"
	"github.com/niflaot/pixels/networking/codec"
)

// Header identifies GROUP_SAVE_BADGE.
const Header uint16 = 1991

const (
	// badgeValuesPerPart is Nitro's element, color, and position tuple width.
	badgeValuesPerPart = 3
	// maxBadgeParts bounds the client-controlled badge allocation.
	maxBadgeParts = 5
)

// Payload contains a group identifier and bounded badge layers.
type Payload struct {
	// GroupID identifies the edited group.
	GroupID int64
	// Parts stores normalized layer triples.
	Parts []grouprecord.BadgePart
}

// Decode validates and unpacks a bounded badge layer collection.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, rest, err := codec.DecodePayload(nil, codec.Definition{codec.Int32Field, codec.Int32Field}, packet.Payload)
	if err != nil {
		return Payload{}, err
	}
	valueCount := int(values[1].Int32)
	if valueCount < badgeValuesPerPart || valueCount > maxBadgeParts*badgeValuesPerPart || valueCount%badgeValuesPerPart != 0 {
		return Payload{}, codec.ErrInvalidField
	}
	partCount := valueCount / badgeValuesPerPart
	result := Payload{GroupID: int64(values[0].Int32), Parts: make([]grouprecord.BadgePart, 0, partCount)}
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
