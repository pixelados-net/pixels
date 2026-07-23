package init

import (
	"context"

	navmodel "github.com/niflaot/pixels/internal/realm/navigator/record"
	roommodel "github.com/niflaot/pixels/internal/realm/room/record/model"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	outevents "github.com/niflaot/pixels/networking/outbound/navigator/browse/eventcategories"
	outflatcats "github.com/niflaot/pixels/networking/outbound/navigator/browse/flatcats"
	outfavorites "github.com/niflaot/pixels/networking/outbound/navigator/favorite/list"
	outcollapsed "github.com/niflaot/pixels/networking/outbound/navigator/session/collapsed"
	outlifted "github.com/niflaot/pixels/networking/outbound/navigator/session/lifted"
	outmetadata "github.com/niflaot/pixels/networking/outbound/navigator/session/metadata"
	outsaved "github.com/niflaot/pixels/networking/outbound/navigator/session/savedsearches"
	outsettings "github.com/niflaot/pixels/networking/outbound/navigator/session/settings"
)

// initialPackets builds the navigator initialization packet set.
func initialPackets(saved []navmodel.SavedSearch, preference navmodel.Preference, lifted []navmodel.LiftedRoom, favorites []int64, categoryPreferences []navmodel.CategoryPreference, categories []roommodel.Category) ([]codec.Packet, error) {
	packets := make([]codec.Packet, 0, 8)
	builders := []packetBuilder{
		func() (codec.Packet, error) { return outmetadata.Encode(metadataContexts(saved)) },
		func() (codec.Packet, error) { return outlifted.Encode(liftedRooms(lifted)) },
		func() (codec.Packet, error) { return outsaved.Encode(savedSearches(saved)) },
		func() (codec.Packet, error) { return outSettings(preference) },
		func() (codec.Packet, error) { return outcollapsed.Encode(collapsedCategories(categoryPreferences)) },
		func() (codec.Packet, error) { return outevents.Encode(nil) },
		func() (codec.Packet, error) { return outfavorites.Encode(FavoriteLimit, int32IDs(favorites)) },
		func() (codec.Packet, error) { return outflatcats.Encode(flatCategories(categories)) },
	}
	for _, builder := range builders {
		packet, err := builder()
		if err != nil {
			return nil, err
		}
		packets = append(packets, packet)
	}

	return packets, nil
}

// sendPackets sends packets in protocol order.
func sendPackets(ctx context.Context, connection netconn.Context, packets []codec.Packet) error {
	for _, packet := range packets {
		if err := connection.Send(ctx, packet); err != nil {
			return err
		}
	}

	return nil
}

// collapsedCategories maps persisted category display state.
func collapsedCategories(preferences []navmodel.CategoryPreference) []string {
	categories := make([]string, 0, len(preferences))
	for _, preference := range preferences {
		if preference.Collapsed {
			categories = append(categories, preference.Code)
		}
	}

	return categories
}

// metadataSearches maps saved searches for one context.
func metadataSearches(searches []navmodel.SavedSearch, code string) []outmetadata.SavedSearch {
	results := make([]outmetadata.SavedSearch, 0, len(searches))
	for _, search := range searches {
		if search.Code == code {
			results = append(results, outmetadata.SavedSearch{ID: int32(search.ID), Code: search.Code, Filter: search.Filter, Localization: search.Localization})
		}
	}

	return results
}

// savedSearches maps saved search records.
func savedSearches(searches []navmodel.SavedSearch) []outsaved.Search {
	results := make([]outsaved.Search, 0, len(searches))
	for _, search := range searches {
		results = append(results, outsaved.Search{ID: int32(search.ID), Code: search.Code, Filter: search.Filter, Localization: search.Localization})
	}

	return results
}

// liftedRooms maps lifted room records.
func liftedRooms(rooms []navmodel.LiftedRoom) []outlifted.Room {
	results := make([]outlifted.Room, 0, len(rooms))
	for _, room := range rooms {
		results = append(results, outlifted.Room{RoomID: int32(room.RoomID), AreaID: int32(room.AreaID), Image: room.Image, Caption: room.Caption})
	}

	return results
}

// flatCategories maps room category records.
func flatCategories(categories []roommodel.Category) []outflatcats.Category {
	results := make([]outflatcats.Category, 0, len(categories))
	for _, category := range categories {
		results = append(results, outflatcats.Category{
			ID:                   int32(category.ID),
			Name:                 category.Caption,
			Visible:              category.Visible,
			Automatic:            category.Automatic,
			AutomaticCategoryKey: category.AutomaticKey,
			GlobalCategoryKey:    category.GlobalKey,
			StaffOnly:            category.StaffOnly,
		})
	}

	return results
}

// outSettings maps navigator preferences.
func outSettings(preference navmodel.Preference) (codec.Packet, error) {
	return outsettings.Encode(outsettings.Params{
		WindowX:         int32(preference.WindowX),
		WindowY:         int32(preference.WindowY),
		WindowWidth:     int32(preference.WindowWidth),
		WindowHeight:    int32(preference.WindowHeight),
		LeftPanelHidden: preference.LeftPanelHidden,
		ResultsMode:     int32(preference.ResultsMode),
	})
}

// int32IDs maps database ids to protocol ids.
func int32IDs(ids []int64) []int32 {
	values := make([]int32, 0, len(ids))
	for _, id := range ids {
		values = append(values, int32(id))
	}

	return values
}

// packetBuilder builds one navigator init packet.
type packetBuilder func() (codec.Packet, error)
