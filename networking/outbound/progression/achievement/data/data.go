// Package data encodes shared achievement wire records.
package data

import "github.com/niflaot/pixels/networking/codec"

// Achievement describes Nitro's exact thirteen-field achievement record.
type Achievement struct {
	// ID identifies the achievement definition.
	ID int32
	// Level stores the current level.
	Level int32
	// BadgeCode stores the current or next badge code.
	BadgeCode string
	// ScoreAtStart stores cumulative progress before the current level.
	ScoreAtStart int32
	// ScoreLimit stores the current level threshold.
	ScoreLimit int32
	// LevelRewardPoints stores the current level wallet reward.
	LevelRewardPoints int32
	// RewardPointType identifies the reward wallet.
	RewardPointType int32
	// CurrentPoints stores cumulative player progress.
	CurrentPoints int32
	// FinalLevel reports whether the maximum level is reached.
	FinalLevel bool
	// Category stores the Nitro category.
	Category string
	// Subcategory stores the Nitro subcategory.
	Subcategory string
	// LevelCount stores the maximum level count.
	LevelCount int32
	// DisplayMethod stores Nitro's progress display method.
	DisplayMethod int32
}

// Append appends one exact achievement record.
func Append(payload []byte, achievement Achievement) ([]byte, error) {
	return codec.AppendPayload(payload, codec.Definition{
		codec.Int32Field, codec.Int32Field, codec.StringField, codec.Int32Field, codec.Int32Field,
		codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.BooleanField, codec.StringField,
		codec.StringField, codec.Int32Field, codec.Int32Field,
	}, codec.Int32(achievement.ID), codec.Int32(achievement.Level), codec.String(achievement.BadgeCode),
		codec.Int32(achievement.ScoreAtStart), codec.Int32(achievement.ScoreLimit), codec.Int32(achievement.LevelRewardPoints),
		codec.Int32(achievement.RewardPointType), codec.Int32(achievement.CurrentPoints), codec.Bool(achievement.FinalLevel),
		codec.String(achievement.Category), codec.String(achievement.Subcategory), codec.Int32(achievement.LevelCount), codec.Int32(achievement.DisplayMethod))
}
