package service

import (
	"strings"

	roommodel "github.com/niflaot/pixels/internal/realm/room/model"
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
