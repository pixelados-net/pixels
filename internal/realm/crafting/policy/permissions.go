// Package policy registers crafting administration permission nodes.
package policy

import "github.com/niflaot/pixels/internal/permission"

var (
	// AltarManageAny permits managing all crafting altars and recipes.
	AltarManageAny = permission.RegisterNode("crafting.altar.manage.any", "")
	// RecyclerManageAny permits managing recycler policy and prizes.
	RecyclerManageAny = permission.RegisterNode("crafting.recycler.manage.any", "")
	// PlayerOverrideAny permits changing a player's known recipes.
	PlayerOverrideAny = permission.RegisterNode("crafting.player.override.any", "")
)
