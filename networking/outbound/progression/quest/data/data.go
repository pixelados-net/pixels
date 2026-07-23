// Package data encodes shared quest wire records.
package data

import "github.com/niflaot/pixels/networking/codec"

// Quest describes Nitro's exact sixteen-field quest record.
type Quest struct {
	// CampaignCode identifies the quest campaign.
	CampaignCode string
	// CompletedInCampaign stores completed predecessors.
	CompletedInCampaign int32
	// CampaignCount stores the campaign quest count.
	CampaignCount int32
	// RewardCurrencyType identifies the reward wallet.
	RewardCurrencyType int32
	// ID identifies the quest.
	ID int32
	// Accepted reports active state.
	Accepted bool
	// Type stores the goal type.
	Type string
	// ImageVersion stores optional client art version.
	ImageVersion string
	// RewardAmount stores the reward amount.
	RewardAmount int32
	// LocalizationCode stores the client localization suffix.
	LocalizationCode string
	// CompletedSteps stores current progress.
	CompletedSteps int32
	// TotalSteps stores required progress.
	TotalSteps int32
	// SortOrder stores client ordering.
	SortOrder int32
	// CatalogPage stores an optional catalog destination.
	CatalogPage string
	// ChainCode stores an optional quest chain code.
	ChainCode string
	// Easy reports daily difficulty.
	Easy bool
}

// Append appends one exact quest record.
func Append(payload []byte, quest Quest) ([]byte, error) {
	return codec.AppendPayload(payload, codec.Definition{
		codec.StringField, codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.BooleanField,
		codec.StringField, codec.StringField, codec.Int32Field, codec.StringField, codec.Int32Field, codec.Int32Field,
		codec.Int32Field, codec.StringField, codec.StringField, codec.BooleanField,
	}, codec.String(quest.CampaignCode), codec.Int32(quest.CompletedInCampaign), codec.Int32(quest.CampaignCount),
		codec.Int32(quest.RewardCurrencyType), codec.Int32(quest.ID), codec.Bool(quest.Accepted), codec.String(quest.Type),
		codec.String(quest.ImageVersion), codec.Int32(quest.RewardAmount), codec.String(quest.LocalizationCode),
		codec.Int32(quest.CompletedSteps), codec.Int32(quest.TotalSteps), codec.Int32(quest.SortOrder), codec.String(quest.CatalogPage),
		codec.String(quest.ChainCode), codec.Bool(quest.Easy))
}
