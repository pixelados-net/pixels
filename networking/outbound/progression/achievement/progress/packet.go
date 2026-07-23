// Package progress encodes ACHIEVEMENT_PROGRESSED responses.
package progress

import (
	"github.com/niflaot/pixels/networking/codec"
	achievementdata "github.com/niflaot/pixels/networking/outbound/progression/achievement/data"
)

// Header identifies ACHIEVEMENT_PROGRESSED.
const Header uint16 = 2107

// Encode creates one incremental achievement progress response.
func Encode(achievement achievementdata.Achievement) (codec.Packet, error) {
	payload, err := achievementdata.Append(nil, achievement)
	return codec.Packet{Header: Header, Payload: payload}, err
}
