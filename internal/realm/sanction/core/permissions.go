package core

import "github.com/niflaot/pixels/internal/permission"

var (
	// ApplyNode permits non-ban global punishments.
	ApplyNode = permission.RegisterNode("moderation.sanction.apply", "")
	// BanNode permits global bans.
	BanNode = permission.RegisterNode("moderation.sanction.ban", "")
	// ImmuneNode protects a player from global staff punishments.
	ImmuneNode = permission.RegisterNode("moderation.sanction.immune", "")
)
