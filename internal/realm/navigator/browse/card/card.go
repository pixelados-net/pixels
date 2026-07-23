// Package projection maps realm records into navigator packet projections.
package projection

import (
	roommodel "github.com/niflaot/pixels/internal/realm/room/record/model"
	"github.com/niflaot/pixels/networking/outbound/navigator/browse/roomcard"
)

// RoomCard maps a room record into a navigator room card.
func RoomCard(room roommodel.Room, userCount int, ranking int, tags []string) roomcard.Card {
	return roomcard.Card{
		RoomID:       int32(room.ID),
		RoomName:     room.Name,
		OwnerID:      int32(room.OwnerPlayerID),
		OwnerName:    room.OwnerName,
		DoorMode:     int32(room.DoorMode),
		UserCount:    int32(userCount),
		MaxUserCount: int32(room.MaxUsers),
		Description:  room.Description,
		TradeMode:    int32(room.TradeMode),
		Score:        int32(room.Score),
		Ranking:      int32(ranking),
		CategoryID:   categoryID(room.CategoryID),
		Tags:         append([]string(nil), tags...),
		ShowOwner:    true,
		AllowPets:    room.AllowPets,
	}
}

// categoryID maps optional category ids to protocol integers.
func categoryID(id *int64) int32 {
	if id == nil {
		return 0
	}

	return int32(*id)
}
