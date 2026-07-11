package service

import (
	"strings"

	roommodel "github.com/niflaot/pixels/internal/realm/room/record/model"
)

// validateCreate validates room creation input.
func validateCreate(params CreateParams) error {
	if params.OwnerPlayerID <= 0 || params.OwnerName == "" {
		return ErrInvalidOwner
	}
	if len(params.Name) < MinRoomNameLength || len(params.Name) > MaxRoomNameLength {
		return ErrInvalidRoomName
	}
	if len(params.Description) > MaxRoomDescriptionLength {
		return ErrInvalidDescription
	}
	if params.MaxUsers <= 0 || params.MaxUsers > MaxRoomUsers {
		return ErrInvalidMaxUsers
	}
	if params.TradeMode < roommodel.TradeModeDisabled || params.TradeMode > roommodel.TradeModeAllowed {
		return ErrInvalidTradeMode
	}
	if params.ModelName == "" {
		return ErrLayoutNotAvailable
	}

	return nil
}

// applyUpdate applies optional settings fields to a room snapshot.
func applyUpdate(room *roommodel.Room, params UpdateParams) {
	if params.Name != nil {
		room.Name = *params.Name
	}
	if params.Description != nil {
		room.Description = *params.Description
	}
	if params.CategoryID != nil {
		room.CategoryID = *params.CategoryID
	}
	if params.MaxUsers != nil {
		room.MaxUsers = *params.MaxUsers
	}
	if params.DoorMode != nil {
		room.DoorMode = *params.DoorMode
	}
	if params.TradeMode != nil {
		room.TradeMode = *params.TradeMode
	}
	if params.AllowWalkthrough != nil {
		room.AllowWalkthrough = *params.AllowWalkthrough
	}
	if params.AllowPets != nil {
		room.AllowPets = *params.AllowPets
	}
	if params.AllowPetsEat != nil {
		room.AllowPetsEat = *params.AllowPetsEat
	}
	if params.HideWalls != nil {
		room.HideWalls = *params.HideWalls
	}
	if params.WallThickness != nil {
		room.WallThickness = *params.WallThickness
	}
	if params.FloorThickness != nil {
		room.FloorThickness = *params.FloorThickness
	}
	if params.ChatMode != nil {
		room.ChatMode = *params.ChatMode
	}
	if params.ChatWeight != nil {
		room.ChatWeight = *params.ChatWeight
	}
	if params.ChatSpeed != nil {
		room.ChatSpeed = *params.ChatSpeed
	}
	if params.ChatDistance != nil {
		room.ChatDistance = *params.ChatDistance
	}
	if params.ChatProtection != nil {
		room.ChatProtection = *params.ChatProtection
	}
	if params.ModerationMute != nil {
		room.ModerationMute = *params.ModerationMute
	}
	if params.ModerationKick != nil {
		room.ModerationKick = *params.ModerationKick
	}
	if params.ModerationBan != nil {
		room.ModerationBan = *params.ModerationBan
	}
}

// validateUpdate validates a merged room settings snapshot.
func validateUpdate(room roommodel.Room, params UpdateParams, tags []string) error {
	if len(room.Name) < MinRoomNameLength || len(room.Name) > MaxRoomNameLength {
		return ErrInvalidRoomName
	}
	if len(room.Description) > MaxRoomDescriptionLength {
		return ErrInvalidDescription
	}
	if room.MaxUsers <= 0 || room.MaxUsers > MaxRoomUsers {
		return ErrInvalidMaxUsers
	}
	if room.DoorMode < roommodel.DoorModeOpen || room.DoorMode > roommodel.DoorModeInvisible {
		return ErrInvalidDoorMode
	}
	if room.TradeMode < roommodel.TradeModeDisabled || room.TradeMode > roommodel.TradeModeAllowed {
		return ErrInvalidTradeMode
	}
	if room.WallThickness < -2 || room.WallThickness > 1 || room.FloorThickness < -2 || room.FloorThickness > 1 {
		return ErrInvalidRoomID
	}
	if room.ChatMode < 0 || room.ChatMode > 2 || room.ChatWeight < 0 || room.ChatWeight > 2 || room.ChatSpeed < 0 || room.ChatSpeed > 2 || room.ChatDistance < 0 || room.ChatDistance > 100 || room.ChatProtection < 0 || room.ChatProtection > 2 {
		return ErrInvalidChatSettings
	}
	if !room.ModerationMute.Valid() || !room.ModerationKick.Valid() || !room.ModerationBan.Valid() {
		return ErrInvalidModerationSettings
	}
	if params.Tags != nil && len(*params.Tags) > MaxRoomTags {
		return ErrInvalidTag
	}
	for _, tag := range tags {
		if len(tag) > MaxRoomTagLength {
			return ErrInvalidTag
		}
		if !params.AllowReservedTags && strings.HasPrefix(tag, "official:") {
			return ErrReservedTag
		}
	}

	return nil
}

// normalizeTags normalizes and bounds room tags.
func normalizeTags(values []string) []string {
	tags := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		tag := strings.ToLower(strings.TrimSpace(value))
		if tag == "" {
			continue
		}
		if _, found := seen[tag]; found {
			continue
		}
		seen[tag] = struct{}{}
		tags = append(tags, tag)
		if len(tags) == MaxRoomTags {
			return tags
		}
	}

	return tags
}
