// Package list encodes ACHIEVEMENT_RESOLUTIONS compatibility responses.
package list

import "github.com/niflaot/pixels/networking/codec"

// Header identifies ACHIEVEMENT_RESOLUTIONS.
const Header uint16 = 66

// Achievement describes one resolution choice.
type Achievement struct {
	// ID identifies the achievement definition.
	ID int32
	// Level stores the required achievement level.
	Level int32
	// BadgeCode stores the required badge.
	BadgeCode string
	// RequiredLevel stores the minimum level.
	RequiredLevel int32
	// State stores the client presentation state.
	State int32
}

// Encode creates one resolution snapshot.
func Encode(itemID int32, achievements []Achievement, endTime int32) (codec.Packet, error) {
	payload, err := codec.AppendPayload(nil, codec.Definition{codec.Int32Field, codec.Int32Field}, codec.Int32(itemID), codec.Int32(int32(len(achievements))))
	for _, achievement := range achievements {
		if err != nil {
			return codec.Packet{}, err
		}
		payload, err = codec.AppendPayload(payload, codec.Definition{codec.Int32Field, codec.Int32Field, codec.StringField, codec.Int32Field, codec.Int32Field}, codec.Int32(achievement.ID), codec.Int32(achievement.Level), codec.String(achievement.BadgeCode), codec.Int32(achievement.RequiredLevel), codec.Int32(achievement.State))
	}
	if err == nil {
		payload, err = codec.AppendPayload(payload, codec.Definition{codec.Int32Field}, codec.Int32(endTime))
	}
	return codec.Packet{Header: Header, Payload: payload}, err
}
