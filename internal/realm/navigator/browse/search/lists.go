package search

import (
	"context"

	outsearch "github.com/niflaot/pixels/networking/outbound/navigator/browse/searchresult"
)

// historyLimit returns a normalized visit-history bound.
func (handler Handler) historyLimit() int {
	if handler.HistoryLimit <= 0 || handler.HistoryLimit > 100 {
		return 25
	}
	return handler.HistoryLimit
}

const (
	// ActionNone tells the client no list action is available.
	ActionNone int32 = 0

	// ActionMore tells the client more results can be requested.
	ActionMore int32 = 1

	// ModeList renders navigator results as rows.
	ModeList int32 = 0

	// ModeThumbnail renders navigator results as thumbnails.
	ModeThumbnail int32 = 1
)

// hotelLists builds the general hotel view.
func (handler Handler) hotelLists(ctx context.Context, query string) ([]outsearch.ResultList, int, error) {
	rooms, err := handler.Rooms.Search(ctx, query, handler.limit())
	if err != nil {
		return nil, 0, err
	}
	cards, err := handler.cards(ctx, rooms)
	if err != nil {
		return nil, 0, err
	}
	list := outsearch.ResultList{Code: "popular", Data: query, Action: ActionMore, Mode: ModeList, Rooms: cards}
	return []outsearch.ResultList{list}, len(list.Rooms), nil
}

// officialLists builds the official view.
func (handler Handler) officialLists(ctx context.Context) ([]outsearch.ResultList, int, error) {
	rooms, err := handler.officialRooms(ctx)
	if err != nil {
		return nil, 0, err
	}

	cards, err := handler.cards(ctx, rooms)
	if err != nil {
		return nil, 0, err
	}
	list := outsearch.ResultList{Code: "official-root", Data: "", Action: ActionNone, Mode: ModeThumbnail, Rooms: cards}
	return []outsearch.ResultList{list}, len(list.Rooms), nil
}

// historyLists builds one recent or frequent room history list.
func (handler Handler) historyLists(ctx context.Context, playerID int64, frequent bool) ([]outsearch.ResultList, int, error) {
	var ids []int64
	var err error
	code := "history"
	if frequent {
		ids, err = handler.Navigator.ListFrequentRoomIDs(ctx, playerID, handler.historyLimit())
		code = "frequent"
	} else {
		ids, err = handler.Navigator.ListRecentRoomIDs(ctx, playerID, handler.historyLimit())
	}
	if err != nil {
		return nil, 0, err
	}
	rooms, err := handler.roomsByIDs(ctx, ids)
	if err != nil {
		return nil, 0, err
	}
	cards, err := handler.cards(ctx, rooms)
	if err != nil {
		return nil, 0, err
	}
	list := outsearch.ResultList{Code: code, Action: ActionNone, Mode: ModeList, Rooms: cards}
	return []outsearch.ResultList{list}, len(cards), nil
}

// eventLists builds the events tab placeholder list.
func (handler Handler) eventLists(ctx context.Context) ([]outsearch.ResultList, int, error) {
	rooms, err := handler.Rooms.ListHighestScore(ctx, handler.limit())
	if err != nil {
		return nil, 0, err
	}

	cards, err := handler.cards(ctx, rooms)
	if err != nil {
		return nil, 0, err
	}
	list := outsearch.ResultList{Code: "categories", Data: "", Action: ActionNone, Mode: ModeList, Rooms: cards}
	return []outsearch.ResultList{list}, len(list.Rooms), nil
}

// myWorldLists builds player-owned and favorite lists.
func (handler Handler) myWorldLists(ctx context.Context, playerID int64) ([]outsearch.ResultList, int, error) {
	owned, err := handler.Rooms.ListByOwner(ctx, playerID)
	if err != nil {
		return nil, 0, err
	}
	favorites, err := handler.favoriteRooms(ctx, playerID)
	if err != nil {
		return nil, 0, err
	}
	ownedCards, err := handler.cards(ctx, owned)
	if err != nil {
		return nil, 0, err
	}
	favoriteCards, err := handler.cards(ctx, favorites)
	if err != nil {
		return nil, 0, err
	}

	lists := []outsearch.ResultList{
		{Code: "my", Data: "", Action: ActionNone, Mode: ModeList, Rooms: ownedCards},
		{Code: "favorites", Data: "", Action: ActionNone, Mode: ModeList, Rooms: favoriteCards},
	}

	return lists, len(lists[0].Rooms) + len(lists[1].Rooms), nil
}

// queryLists builds a direct search result.
func (handler Handler) queryLists(ctx context.Context, code string, query string) ([]outsearch.ResultList, int, error) {
	rooms, err := handler.Rooms.Search(ctx, query, handler.limit())
	if err != nil {
		return nil, 0, err
	}

	cards, err := handler.cards(ctx, rooms)
	if err != nil {
		return nil, 0, err
	}
	list := outsearch.ResultList{Code: code, Data: query, Action: ActionMore, Mode: ModeList, Rooms: cards}
	return []outsearch.ResultList{list}, len(list.Rooms), nil
}
