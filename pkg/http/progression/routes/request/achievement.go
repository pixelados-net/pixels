package request

// AchievementCreate creates one data-driven achievement.
type AchievementCreate struct {
	Audit
	// Name stores the badge group name without the ACH prefix.
	Name string `json:"name"`
	// Category stores the Nitro category.
	Category string `json:"category"`
	// Subcategory stores the optional Nitro subcategory.
	Subcategory string `json:"subcategory"`
	// TriggerKey identifies the gameplay signal.
	TriggerKey string `json:"triggerKey"`
	// Visible optionally controls client listing and defaults true.
	Visible *bool `json:"visible"`
}

// AchievementUpdate edits one definition under optimistic locking.
type AchievementUpdate struct {
	Audit
	// Version stores the expected durable version.
	Version int64 `json:"version"`
	// Category optionally replaces the Nitro category.
	Category *string `json:"category"`
	// Subcategory optionally replaces the Nitro subcategory.
	Subcategory *string `json:"subcategory"`
	// TriggerKey optionally replaces the gameplay signal.
	TriggerKey *string `json:"triggerKey"`
	// Visible optionally controls client listing.
	Visible *bool `json:"visible"`
	// Enabled optionally controls progression fan-out.
	Enabled *bool `json:"enabled"`
}

// AchievementLevel creates or updates one cumulative level.
type AchievementLevel struct {
	Audit
	// Level stores the one-based level number for creation.
	Level int32 `json:"level"`
	// ProgressNeeded stores the cumulative threshold.
	ProgressNeeded int64 `json:"progressNeeded"`
	// RewardCurrencyType identifies the rewarded wallet.
	RewardCurrencyType int32 `json:"rewardCurrencyType"`
	// RewardAmount stores the wallet reward.
	RewardAmount int64 `json:"rewardAmount"`
	// ScorePoints stores achievement score earned.
	ScorePoints int32 `json:"scorePoints"`
}
