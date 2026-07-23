// Package limits encodes BADGE_POINT_LIMITS responses.
package limits

import "github.com/niflaot/pixels/networking/codec"

// Header identifies BADGE_POINT_LIMITS.
const Header uint16 = 2501

// Level describes one badge level suffix and threshold.
type Level struct {
	// Suffix stores the badge level suffix.
	Suffix int32
	// Limit stores the cumulative threshold.
	Limit int32
}

// Group describes one achievement badge prefix and thresholds.
type Group struct {
	// Prefix stores the name after ACH_.
	Prefix string
	// Levels stores ordered thresholds.
	Levels []Level
}

// Encode creates one complete badge point limit snapshot.
func Encode(groups []Group) (codec.Packet, error) {
	payload, err := codec.AppendPayload(nil, codec.Definition{codec.Int32Field}, codec.Int32(int32(len(groups))))
	for _, group := range groups {
		if err != nil {
			return codec.Packet{}, err
		}
		payload, err = codec.AppendPayload(payload, codec.Definition{codec.StringField, codec.Int32Field}, codec.String(group.Prefix), codec.Int32(int32(len(group.Levels))))
		for _, level := range group.Levels {
			if err != nil {
				return codec.Packet{}, err
			}
			payload, err = codec.AppendPayload(payload, codec.Definition{codec.Int32Field, codec.Int32Field}, codec.Int32(level.Suffix), codec.Int32(level.Limit))
		}
	}
	return codec.Packet{Header: Header, Payload: payload}, err
}
