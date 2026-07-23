// Package levelup encodes ACHIEVEMENT_NOTIFICATION responses.
package levelup

import "github.com/niflaot/pixels/networking/codec"

// Header identifies ACHIEVEMENT_NOTIFICATION.
const Header uint16 = 806

// Data describes Nitro's exact twelve-field level-up record.
type Data struct {
	// Type stores the notification type.
	Type int32
	// Level stores the new level.
	Level int32
	// BadgeID identifies the durable badge row.
	BadgeID int32
	// BadgeCode stores the new badge code.
	BadgeCode string
	// Points stores cumulative achievement progress.
	Points int32
	// RewardPoints stores the wallet reward amount.
	RewardPoints int32
	// RewardPointType identifies the reward wallet.
	RewardPointType int32
	// BonusPoints stores score earned by the level.
	BonusPoints int32
	// AchievementID identifies the definition.
	AchievementID int32
	// RemovedBadgeCode stores the replaced badge code.
	RemovedBadgeCode string
	// Category stores the Nitro category.
	Category string
	// ShowDialog controls the client level-up dialog.
	ShowDialog bool
}

// Encode creates one achievement level-up notification.
func Encode(data Data) (codec.Packet, error) {
	return codec.NewPacket(Header, codec.Definition{
		codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.StringField, codec.Int32Field, codec.Int32Field,
		codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.StringField, codec.StringField, codec.BooleanField,
	}, codec.Int32(data.Type), codec.Int32(data.Level), codec.Int32(data.BadgeID), codec.String(data.BadgeCode),
		codec.Int32(data.Points), codec.Int32(data.RewardPoints), codec.Int32(data.RewardPointType), codec.Int32(data.BonusPoints),
		codec.Int32(data.AchievementID), codec.String(data.RemovedBadgeCode), codec.String(data.Category), codec.Bool(data.ShowDialog))
}
