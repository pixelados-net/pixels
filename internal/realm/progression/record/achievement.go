// Package record defines progression persistence records and contracts.
package record

import "time"

// AchievementDefinition describes one data-driven achievement group.
type AchievementDefinition struct {
	// ID identifies the definition.
	ID int64
	// Name stores the badge group name without the ACH prefix.
	Name string
	// Category stores the Nitro category.
	Category string
	// Subcategory stores the optional Nitro subcategory.
	Subcategory string
	// TriggerKey identifies the gameplay signal.
	TriggerKey string
	// Visible controls client listing.
	Visible bool
	// Enabled controls progress fan-out.
	Enabled bool
	// Version stores optimistic administration state.
	Version int64
	// Levels stores ascending cumulative thresholds.
	Levels []AchievementLevel
}

// AchievementLevel describes one cumulative achievement threshold.
type AchievementLevel struct {
	// DefinitionID identifies the parent definition.
	DefinitionID int64
	// Level stores the one-based level number.
	Level int32
	// ProgressNeeded stores the cumulative threshold.
	ProgressNeeded int64
	// RewardCurrencyType identifies the rewarded wallet.
	RewardCurrencyType int32
	// RewardAmount stores the wallet reward.
	RewardAmount int64
	// ScorePoints stores achievement score earned.
	ScorePoints int32
}

// PlayerAchievement stores one player's durable progress row.
type PlayerAchievement struct {
	// PlayerID identifies the player.
	PlayerID int64
	// DefinitionID identifies the achievement definition.
	DefinitionID int64
	// Progress stores cumulative progress.
	Progress int64
	// Level stores the highest paid level.
	Level int32
	// LastDailyAt stores the most recent daily trigger date.
	LastDailyAt *time.Time
}

// ProgressMutation stores one locked progress mutation result.
type ProgressMutation struct {
	// Before stores state before mutation.
	Before PlayerAchievement
	// After stores state after mutation.
	After PlayerAchievement
	// Crossed stores newly crossed levels in ascending order.
	Crossed []AchievementLevel
}
