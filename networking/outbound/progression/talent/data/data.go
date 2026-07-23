// Package data defines shared talent track wire records.
package data

// Task describes one achievement prerequisite projection.
type Task struct {
	// ID identifies the achievement.
	ID int32
	// Index stores the required level index.
	Index int32
	// BadgeCode stores the required badge.
	BadgeCode string
	// State stores locked, active, or completed state.
	State int32
	// Progress stores current achievement progress.
	Progress int32
	// RequiredProgress stores the threshold.
	RequiredProgress int32
}

// Product describes one talent furniture or subscription reward.
type Product struct {
	// Name stores the product code.
	Name string
	// Value stores sprite or VIP-day metadata.
	Value int32
}

// Level describes one nested talent track level.
type Level struct {
	// ID identifies the level.
	ID int32
	// State stores locked, active, or completed state.
	State int32
	// Tasks stores achievement prerequisites.
	Tasks []Task
	// Perks stores Nitro perk names.
	Perks []string
	// Products stores furniture rewards.
	Products []Product
}
