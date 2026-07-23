// Package skilllist encodes BOT_SKILL_LIST_UPDATE.
package skilllist

import "github.com/niflaot/pixels/networking/codec"

// Header identifies BOT_SKILL_LIST_UPDATE.
const Header uint16 = 69

// Skill stores one bot skill and optional data.
type Skill struct {
	// ID identifies the skill.
	ID int32
	// Data stores skill-specific state.
	Data string
}

// Encode creates BOT_SKILL_LIST_UPDATE.
func Encode(botID int64, skills []Skill) (codec.Packet, error) {
	payload, err := codec.AppendPayload(nil, codec.Definition{codec.Int32Field, codec.Int32Field}, codec.Int32(int32(-botID)), codec.Int32(int32(len(skills))))
	if err != nil {
		return codec.Packet{}, err
	}
	for _, skill := range skills {
		payload, err = codec.AppendPayload(payload, codec.Definition{codec.Int32Field, codec.StringField}, codec.Int32(skill.ID), codec.String(skill.Data))
		if err != nil {
			return codec.Packet{}, err
		}
	}
	return codec.Packet{Header: Header, Payload: payload}, nil
}
