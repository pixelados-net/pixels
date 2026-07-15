package routes

import (
	roommodel "github.com/niflaot/pixels/internal/realm/room/record/model"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
)

// roomResponses maps room records.
func roomResponses(rooms []roommodel.Room) []RoomResponse {
	items := make([]RoomResponse, 0, len(rooms))
	for _, room := range rooms {
		items = append(items, roomResponse(room))
	}

	return items
}

// roomResponse maps one room record.
func roomResponse(room roommodel.Room) RoomResponse {
	return RoomResponse{
		ID:               room.ID,
		Version:          room.Version.Version,
		Name:             room.Name,
		OwnerPlayerID:    room.OwnerPlayerID,
		OwnerName:        room.OwnerName,
		ModelName:        room.ModelName,
		MaxUsers:         room.MaxUsers,
		CategoryID:       room.CategoryID,
		Score:            room.Score,
		IsBundleTemplate: room.IsBundleTemplate,
		RollerSpeed:      room.RollerSpeed,
	}
}

// occupancyResponse maps runtime occupancy.
func occupancyResponse(occupancy roomlive.Occupancy) OccupancyResponse {
	return OccupancyResponse{
		RoomID:    occupancy.RoomID,
		Count:     occupancy.Count,
		MaxUsers:  occupancy.MaxUsers,
		PlayerIDs: append([]int64(nil), occupancy.PlayerIDs...),
	}
}

// categoryResponses maps room categories.
func categoryResponses(categories []roommodel.Category) []CategoryResponse {
	items := make([]CategoryResponse, 0, len(categories))
	for _, category := range categories {
		items = append(items, CategoryResponse{
			ID:      category.ID,
			Caption: category.Caption,
			Visible: category.Visible,
			Order:   category.Order,
		})
	}

	return items
}
