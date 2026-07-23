// Package policy declares progression permission capabilities.
package policy

import "github.com/niflaot/pixels/internal/permission"

var (
	// ManageDefinitions permits progression catalog administration.
	ManageDefinitions = permission.RegisterNode("progression.definitions.manage.any", "")
	// OverridePlayers permits player progression and badge overrides.
	OverridePlayers = permission.RegisterNode("progression.player.override.any", "")
	// ManageQuests permits quest, quiz, and promo administration.
	ManageQuests = permission.RegisterNode("progression.quest.manage.any", "")
	// TradePerk exposes Nitro's first real TRADE perk.
	TradePerk = permission.RegisterNode("progression.perk.trade", "TRADE")
)
