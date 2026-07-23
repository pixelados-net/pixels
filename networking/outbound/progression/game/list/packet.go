// Package list encodes GAME_ACHIEVEMENTS compatibility responses.
package list

import "github.com/niflaot/pixels/networking/codec"

// Header identifies GAME_ACHIEVEMENTS.
const Header uint16 = 1689

// Achievement describes one game achievement.
type Achievement struct {
	// ID identifies the achievement.
	ID int32
	// Name stores the achievement code.
	Name string
	// Levels stores the available level count.
	Levels int32
}

// Group describes achievements belonging to one game type.
type Group struct {
	// GameTypeID identifies the game.
	GameTypeID int32
	// Achievements stores the game's definitions.
	Achievements []Achievement
}

// Encode creates one grouped game achievement snapshot.
func Encode(groups []Group) (codec.Packet, error) {
	payload, err := codec.AppendPayload(nil, codec.Definition{codec.Int32Field}, codec.Int32(int32(len(groups))))
	for _, group := range groups {
		if err != nil {
			return codec.Packet{}, err
		}
		payload, err = codec.AppendPayload(payload, codec.Definition{codec.Int32Field, codec.Int32Field}, codec.Int32(group.GameTypeID), codec.Int32(int32(len(group.Achievements))))
		for _, achievement := range group.Achievements {
			if err != nil {
				return codec.Packet{}, err
			}
			payload, err = codec.AppendPayload(payload, codec.Definition{codec.Int32Field, codec.StringField, codec.Int32Field}, codec.Int32(achievement.ID), codec.String(achievement.Name), codec.Int32(achievement.Levels))
		}
	}
	return codec.Packet{Header: Header, Payload: payload}, err
}
