package info

import (
	"context"

	roommodel "github.com/niflaot/pixels/internal/realm/room/record/model"
	outinfo "github.com/niflaot/pixels/networking/outbound/navigator/browse/roominfo"
)

// roomTags loads room tags as protocol strings.
func (handler Handler) roomTags(ctx context.Context, roomID int64) ([]string, error) {
	tags, err := handler.Rooms.ListTags(ctx, roomID)
	if err != nil {
		return nil, err
	}

	values := make([]string, 0, len(tags))
	for _, tag := range tags {
		values = append(values, tag.Value)
	}

	return values, nil
}

// userCount returns live occupancy for a room.
func (handler Handler) userCount(roomID int64) int {
	if handler.Runtime == nil {
		return 0
	}
	active, found := handler.Runtime.Find(roomID)
	if !found {
		return 0
	}

	return active.Occupancy().Count
}

// moderation maps room moderation settings.
func moderation(room roommodel.Room) outinfo.ModerationSettings {
	return outinfo.ModerationSettings{
		AllowMute: int32(room.ModerationMute),
		AllowKick: int32(room.ModerationKick),
		AllowBan:  int32(room.ModerationBan),
	}
}

// chat maps room chat settings.
func chat(room roommodel.Room) outinfo.ChatSettings {
	return outinfo.ChatSettings{
		Mode:       int32(room.ChatMode),
		Weight:     int32(room.ChatWeight),
		Speed:      int32(room.ChatSpeed),
		Distance:   int32(room.ChatDistance),
		Protection: int32(room.ChatProtection),
	}
}
