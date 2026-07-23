// Package list encodes the ACHIEVEMENT_LIST outbound snapshot.
package list

import (
	"github.com/niflaot/pixels/networking/codec"
	achievementdata "github.com/niflaot/pixels/networking/outbound/progression/achievement/data"
)

// Header identifies ACHIEVEMENT_LIST.
const Header uint16 = 305

// Encode creates one complete achievement snapshot.
func Encode(achievements []achievementdata.Achievement, defaultCategory string) (codec.Packet, error) {
	payload, err := codec.AppendPayload(nil, codec.Definition{codec.Int32Field}, codec.Int32(int32(len(achievements))))
	for _, achievement := range achievements {
		if err != nil {
			return codec.Packet{}, err
		}
		payload, err = achievementdata.Append(payload, achievement)
	}
	if err == nil {
		payload, err = codec.AppendPayload(payload, codec.Definition{codec.StringField}, codec.String(defaultCategory))
	}
	return codec.Packet{Header: Header, Payload: payload}, err
}
