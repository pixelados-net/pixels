package save

import (
	roommodel "github.com/niflaot/pixels/internal/realm/room/model"
	roomservice "github.com/niflaot/pixels/internal/realm/room/service"
)

// updateParams maps protocol input to service update fields.
func updateParams(input Command, allowReserved bool) roomservice.UpdateParams {
	doorMode := roommodel.DoorMode(input.DoorMode)
	tradeMode := roommodel.TradeMode(input.TradeMode)
	mute := roommodel.ModerationPolicy(input.ModerationMute)
	kick := roommodel.ModerationPolicy(input.ModerationKick)
	ban := roommodel.ModerationPolicy(input.ModerationBan)
	categoryID := &input.CategoryID
	if input.CategoryID <= 0 {
		categoryID = nil
	}
	params := roomservice.UpdateParams{Name: &input.Name, Description: &input.Description, CategoryID: &categoryID,
		Tags: &input.Tags, MaxUsers: &input.MaxUsers, DoorMode: &doorMode, TradeMode: &tradeMode,
		AllowWalkthrough: &input.AllowWalkthrough, AllowPets: &input.AllowPets, AllowPetsEat: &input.AllowPetsEat,
		HideWalls: &input.HideWalls, WallThickness: &input.WallThickness, FloorThickness: &input.FloorThickness,
		ChatMode: &input.ChatMode, ChatWeight: &input.ChatWeight, ChatSpeed: &input.ChatSpeed,
		ChatDistance: &input.ChatDistance, ChatProtection: &input.ChatProtection,
		ModerationMute: &mute, ModerationKick: &kick, ModerationBan: &ban, AllowReservedTags: allowReserved}
	if input.Password != "" {
		params.Password = &input.Password
	}

	return params
}

// clubFieldsChanged reports whether one save mutates HC-restricted rendering settings.
func clubFieldsChanged(room roommodel.Room, input Command) bool {
	return room.HideWalls != input.HideWalls || room.WallThickness != input.WallThickness || room.FloorThickness != input.FloorThickness
}

// clubFieldsAllowed applies entitlement and global-management policy to HC fields.
func clubFieldsAllowed(room roommodel.Room, input Command, hasClub bool, globalManager bool) bool {
	return !clubFieldsChanged(room, input) || hasClub || globalManager
}

// moderationFieldsChanged reports whether one save mutates room moderation policy.
func moderationFieldsChanged(room roommodel.Room, input Command) bool {
	return room.ModerationMute != roommodel.ModerationPolicy(input.ModerationMute) ||
		room.ModerationKick != roommodel.ModerationPolicy(input.ModerationKick) ||
		room.ModerationBan != roommodel.ModerationPolicy(input.ModerationBan)
}
