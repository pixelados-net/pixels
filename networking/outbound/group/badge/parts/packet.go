// Package parts contains GROUP_BADGE_PARTS.
package parts

import (
	"github.com/niflaot/pixels/internal/realm/group/badge"
	grouprecord "github.com/niflaot/pixels/internal/realm/group/record"
	"github.com/niflaot/pixels/networking/codec"
)

// Header identifies GROUP_BADGE_PARTS.
const Header uint16 = 2238

// Encode creates all five Nitro badge editor collections.
func Encode(snapshot *badge.Snapshot) (codec.Packet, error) {
	if snapshot == nil {
		return codec.Packet{}, codec.ErrInvalidField
	}
	payload, err := appendElements(nil, snapshot.Elements, grouprecord.BadgeBase)
	if err != nil {
		return codec.Packet{}, err
	}
	payload, err = appendElements(payload, snapshot.Elements, grouprecord.BadgeSymbol)
	if err != nil {
		return codec.Packet{}, err
	}
	for _, family := range []grouprecord.ColorFamily{grouprecord.BaseColor, grouprecord.SymbolColor, grouprecord.BackgroundColor} {
		payload, err = appendColors(payload, snapshot.Colors, family)
		if err != nil {
			return codec.Packet{}, err
		}
	}
	return codec.Packet{Header: Header, Payload: payload}, nil
}

// appendElements appends one filtered element collection.
func appendElements(dst []byte, elements []grouprecord.BadgeElement, kind grouprecord.BadgeKind) ([]byte, error) {
	count := 0
	for _, element := range elements {
		if element.Kind == kind {
			count++
		}
	}
	dst, err := codec.AppendPayload(dst, codec.Definition{codec.Int32Field}, codec.Int32(int32(count)))
	if err != nil {
		return nil, err
	}
	for _, element := range elements {
		if element.Kind != kind {
			continue
		}
		dst, err = codec.AppendPayload(dst, codec.Definition{codec.Int32Field, codec.StringField, codec.StringField}, codec.Int32(element.ID), codec.String(element.ValueA), codec.String(element.ValueB))
		if err != nil {
			return nil, err
		}
	}
	return dst, nil
}

// appendColors appends one filtered color collection.
func appendColors(dst []byte, colors []grouprecord.BadgeColor, family grouprecord.ColorFamily) ([]byte, error) {
	count := 0
	for _, color := range colors {
		if color.Family == family {
			count++
		}
	}
	dst, err := codec.AppendPayload(dst, codec.Definition{codec.Int32Field}, codec.Int32(int32(count)))
	if err != nil {
		return nil, err
	}
	for _, color := range colors {
		if color.Family != family {
			continue
		}
		dst, err = codec.AppendPayload(dst, codec.Definition{codec.Int32Field, codec.StringField}, codec.Int32(color.ID), codec.String(color.Hex))
		if err != nil {
			return nil, err
		}
	}
	return dst, nil
}
