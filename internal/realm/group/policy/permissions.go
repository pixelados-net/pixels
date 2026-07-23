// Package policy registers social-group hotel permission nodes.
package policy

import "github.com/niflaot/pixels/internal/permission"

var (
	// CreateNode permits player-originated social-group creation.
	CreateNode = permission.RegisterNode("group.create", "")
	// ManageAny permits editing any active social group.
	ManageAny = permission.RegisterNode("group.manage.any", "")
	// DeleteAny permits deactivating any social group.
	DeleteAny = permission.RegisterNode("group.delete.any", "")
	// MembersManageAny permits managing any social-group roster.
	MembersManageAny = permission.RegisterNode("group.members.manage.any", "")
	// RolesManageAny permits managing any social-group roles.
	RolesManageAny = permission.RegisterNode("group.roles.manage.any", "")
	// HomeRoomRebind permits administrative home-room replacement.
	HomeRoomRebind = permission.RegisterNode("group.home_room.rebind", "")
	// BadgeManageAny permits editing any social-group badge.
	BadgeManageAny = permission.RegisterNode("group.badge.manage.any", "")
	// ForumManageAny permits managing any group forum.
	ForumManageAny = permission.RegisterNode("group.forum.manage.any", "")
	// ForumModerateAny permits moderating any group forum.
	ForumModerateAny = permission.RegisterNode("group.forum.moderate.any", "")
	// ReadDeactivated permits reading retained deactivated group data.
	ReadDeactivated = permission.RegisterNode("group.read.deactivated", "")
)
