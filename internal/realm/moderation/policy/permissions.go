// Package policy defines global moderation capability nodes.
package policy

import "github.com/niflaot/pixels/internal/permission"

var (
	// ToolAccess permits opening the global moderator tool.
	ToolAccess = permission.RegisterNode("moderation.tool.access", "")
	// IssueManage permits claiming and resolving issues.
	IssueManage = permission.RegisterNode("moderation.issue.manage", "")
	// ChatlogRead permits reading evidence and room visits.
	ChatlogRead = permission.RegisterNode("moderation.chatlog.read", "")
	// RoomOverride permits global room warnings and setting overrides.
	RoomOverride = permission.RegisterNode("moderation.room.override", "")
	// GuideDuty permits guide pool participation.
	GuideDuty = permission.RegisterNode("moderation.guide.duty", "")
	// GuardianDuty permits guardian review participation.
	GuardianDuty = permission.RegisterNode("moderation.guardian.duty", "")
)
