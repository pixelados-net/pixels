package room

import "github.com/niflaot/pixels/internal/permission"

var (
	// EnterAny allows a player to enter regardless of room access mode or ban.
	EnterAny = permission.RegisterNode("room.enter.any", "")

	// EnterFull allows a player to enter a room at its normal capacity.
	EnterFull = permission.RegisterNode("room.enter.full", "")

	// AnswerAnyDoorbell allows a player to answer doorbells in any occupied room.
	AnswerAnyDoorbell = permission.RegisterNode("room.doorbell.answer.any", "")

	// ModerationOwnKick allows kicking from rooms governed by local policy.
	ModerationOwnKick = permission.RegisterNode("room.moderation.own.kick", "")
	// ModerationOwnMute allows muting in rooms governed by local policy.
	ModerationOwnMute = permission.RegisterNode("room.moderation.own.mute", "")
	// ModerationOwnBan allows banning in rooms governed by local policy.
	ModerationOwnBan = permission.RegisterNode("room.moderation.own.ban", "")
	// ModerationAnyKick allows staff to kick from any room.
	ModerationAnyKick = permission.RegisterNode("room.moderation.any.kick", "")
	// ModerationAnyMute allows staff to mute in any room.
	ModerationAnyMute = permission.RegisterNode("room.moderation.any.mute", "")
	// ModerationAnyBan allows staff to ban in any room.
	ModerationAnyBan = permission.RegisterNode("room.moderation.any.ban", "")
	// RightsOwnGrant allows owners to grant build rights.
	RightsOwnGrant = permission.RegisterNode("room.rights.own.grant", "")
	// RightsOwnRevoke allows owners to revoke build rights.
	RightsOwnRevoke = permission.RegisterNode("room.rights.own.revoke", "")
	// RightsAnyGrant allows staff to grant build rights in any room.
	RightsAnyGrant = permission.RegisterNode("room.rights.any.grant", "")
	// RightsAnyRevoke allows staff to revoke build rights in any room.
	RightsAnyRevoke = permission.RegisterNode("room.rights.any.revoke", "")
	// Unkickable protects a player from room moderation actions.
	Unkickable = permission.RegisterNode("room.unkickable", "")
)
