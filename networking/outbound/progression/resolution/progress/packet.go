// Package progress encodes ACHIEVEMENT_RESOLUTION_PROGRESS responses.
package progress

import "github.com/niflaot/pixels/networking/codec"

// Header identifies ACHIEVEMENT_RESOLUTION_PROGRESS.
const Header uint16 = 3370

// Encode creates one resolution progress response.
func Encode(itemID int32, achievementID int32, badgeCode string, playerProgress int32, totalProgress int32, endTime int32) (codec.Packet, error) {
	definition := codec.Definition{codec.Int32Field, codec.Int32Field, codec.StringField, codec.Int32Field, codec.Int32Field, codec.Int32Field}
	return codec.NewPacket(Header, definition, codec.Int32(itemID), codec.Int32(achievementID), codec.String(badgeCode), codec.Int32(playerProgress), codec.Int32(totalProgress), codec.Int32(endTime))
}
