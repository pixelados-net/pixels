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
	// SettingsOwnManage allows owners and rights holders to manage local room settings.
	SettingsOwnManage = permission.RegisterNode("room.settings.own.manage", "")
	// SettingsAnyManage allows staff to manage settings in any room.
	SettingsAnyManage = permission.RegisterNode("room.settings.any.manage", "")
	// ModerationOwnPolicyManage allows owners to configure room moderation policy.
	ModerationOwnPolicyManage = permission.RegisterNode("room.moderation.policy.own.manage", "")
	// ModerationAnyPolicyManage allows staff to configure moderation policy in any room.
	ModerationAnyPolicyManage = permission.RegisterNode("room.moderation.policy.any.manage", "")
	// FloorplanOwnEdit allows local owners and rights holders to edit floor plans.
	FloorplanOwnEdit = permission.RegisterNode("room.floorplan.own.edit", "")
	// FloorplanAnyEdit allows staff to edit any room floor plan.
	FloorplanAnyEdit = permission.RegisterNode("room.floorplan.any.edit", "")
	// BundleTemplateManage allows administration of catalog room templates.
	BundleTemplateManage = permission.RegisterNode("room.admin.bundle_template.manage", "")
	// WiredConfigure identifies ordinary owner and rights-holder WIRED configuration.
	WiredConfigure = permission.RegisterNode("room.wired.configure", "")
	// WiredConfigureAny allows staff to configure WIRED in any active room.
	WiredConfigureAny = permission.RegisterNode("room.wired.configure.any", "")
	// WiredInspect allows staff to inspect configurations and runtime traces.
	WiredInspect = permission.RegisterNode("room.wired.inspect", "")
	// WiredAdmin allows protected WIRED HTTP mutations and reloads.
	WiredAdmin = permission.RegisterNode("room.wired.admin", "")
	// WiredRewardManage allows editing durable WIRED reward definitions.
	WiredRewardManage = permission.RegisterNode("room.wired.reward.manage", "")
	// WiredCompatibilityUse allows API-only compatibility behaviors.
	WiredCompatibilityUse = permission.RegisterNode("room.wired.compatibility.use", "")
	// PromotionManageAny allows staff to purchase or edit promotions for any room.
	PromotionManageAny = permission.RegisterNode("room.promotion.manage.any", "")
	// DeleteAny allows staff to delete any room through the native settings action.
	DeleteAny = permission.RegisterNode("room.delete.any", "")
	// StaffPickManage allows staff to add rooms to the official Navigator selection.
	StaffPickManage = permission.RegisterNode("room.staffpick.manage", "")
	// AmbassadorAlert allows a player to submit the native in-room ambassador alert.
	AmbassadorAlert = permission.RegisterNode("room.ambassador.alert", "")
)
