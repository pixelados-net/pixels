// Package friendnotification encodes MESSENGER_FRIEND_NOTIFICATION responses.
package friendnotification

import "github.com/niflaot/pixels/networking/codec"

// Header identifies MESSENGER_FRIEND_NOTIFICATION.
const Header uint16 = 3082

// TypeAchievementCompleted identifies an achievement level-up notification.
const TypeAchievementCompleted int32 = 1

// TypeQuestCompleted identifies a quest completion notification.
const TypeQuestCompleted int32 = 2

// Encode creates one friend toolbar notification.
func Encode(playerID string, notificationType int32, data string) (codec.Packet, error) {
	return codec.NewPacket(Header, codec.Definition{codec.StringField, codec.Int32Field, codec.StringField}, codec.String(playerID), codec.Int32(notificationType), codec.String(data))
}
