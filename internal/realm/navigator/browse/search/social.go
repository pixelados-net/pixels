package search

import (
	"context"
	"strconv"

	roommodel "github.com/niflaot/pixels/internal/realm/room/record/model"
	roomservice "github.com/niflaot/pixels/internal/realm/room/record/service"
	outsearch "github.com/niflaot/pixels/networking/outbound/navigator/browse/searchresult"
)

// roomList projects one named room result list.
func (handler Handler) roomList(ctx context.Context, code string, rooms []roommodel.Room) ([]outsearch.ResultList, int, error) {
	cards, err := handler.cards(ctx, rooms)
	if err != nil {
		return nil, 0, err
	}
	list := outsearch.ResultList{Code: code, Action: ActionNone, Mode: ModeList, Rooms: cards}
	return []outsearch.ResultList{list}, len(cards), nil
}

// ownedLists builds rooms owned by the actor.
func (handler Handler) ownedLists(ctx context.Context, playerID int64) ([]outsearch.ResultList, int, error) {
	rooms, err := handler.Rooms.ListByOwner(ctx, playerID)
	if err != nil {
		return nil, 0, err
	}
	return handler.roomList(ctx, "my", rooms)
}

// favoriteLists builds the actor's visible favorite rooms.
func (handler Handler) favoriteLists(ctx context.Context, playerID int64) ([]outsearch.ResultList, int, error) {
	rooms, err := handler.favoriteRooms(ctx, playerID)
	if err != nil {
		return nil, 0, err
	}
	return handler.roomList(ctx, "favorites", rooms)
}

// rightsLists builds rooms where the actor holds explicit rights.
func (handler Handler) rightsLists(ctx context.Context, playerID int64) ([]outsearch.ResultList, int, error) {
	if handler.RightRooms == nil {
		return handler.roomList(ctx, "rights", nil)
	}
	ids, err := handler.RightRooms.RoomIDsForPlayer(ctx, playerID)
	if err != nil {
		return nil, 0, err
	}
	rooms, err := handler.roomsByIDs(ctx, ids)
	if err != nil {
		return nil, 0, err
	}
	return handler.roomList(ctx, "rights", rooms)
}

// friendCurrentLists builds unique active rooms occupied by friends.
func (handler Handler) friendCurrentLists(ctx context.Context, playerID int64) ([]outsearch.ResultList, int, error) {
	ids, err := handler.friendIDs(ctx, playerID)
	if err != nil {
		return nil, 0, err
	}
	roomIDs := make([]int64, 0, len(ids))
	seen := make(map[int64]struct{}, len(ids))
	for _, id := range ids {
		friend, found := handler.Players.Find(id)
		if !found {
			continue
		}
		roomID, found := friend.CurrentRoom()
		if _, exists := seen[roomID]; found && !exists {
			seen[roomID] = struct{}{}
			roomIDs = append(roomIDs, roomID)
		}
	}
	rooms, err := handler.roomsByIDs(ctx, roomIDs)
	if err != nil {
		return nil, 0, err
	}
	return handler.roomList(ctx, "friends_now", rooms)
}

// friendOwnedLists builds rooms owned by the actor's friends.
func (handler Handler) friendOwnedLists(ctx context.Context, playerID int64) ([]outsearch.ResultList, int, error) {
	ids, err := handler.friendIDs(ctx, playerID)
	if err != nil {
		return nil, 0, err
	}
	rooms := make([]roommodel.Room, 0)
	if finder, ok := handler.Rooms.(roomservice.OwnerBatchFinder); ok {
		rooms, err = finder.ListByOwnerIDs(ctx, ids)
	} else {
		for _, id := range ids {
			owned, listErr := handler.Rooms.ListByOwner(ctx, id)
			if listErr != nil {
				return nil, 0, listErr
			}
			rooms = append(rooms, owned...)
		}
	}
	if err != nil {
		return nil, 0, err
	}
	return handler.roomList(ctx, "friends_rooms", rooms)
}

// friendIDs reads the actor's directional friend identifiers.
func (handler Handler) friendIDs(ctx context.Context, playerID int64) ([]int64, error) {
	if handler.Friends == nil {
		return []int64{}, nil
	}
	return handler.Friends.FriendIDs(ctx, playerID)
}

// groupMembershipLists builds headquarters for active memberships.
func (handler Handler) groupMembershipLists(ctx context.Context, playerID int64) ([]outsearch.ResultList, int, error) {
	if handler.GroupRooms == nil {
		return handler.roomList(ctx, "my_groups", nil)
	}
	groups, err := handler.GroupRooms.PlayerGroups(ctx, playerID)
	if err != nil {
		return nil, 0, err
	}
	ids := make([]int64, 0, len(groups))
	for _, group := range groups {
		if group.Group.Active() && group.Group.HomeRoomID > 0 {
			ids = append(ids, group.Group.HomeRoomID)
		}
	}
	rooms, err := handler.roomsByIDs(ctx, ids)
	if err != nil {
		return nil, 0, err
	}
	return handler.roomList(ctx, "my_groups", rooms)
}

// popularGroupLists builds headquarters ordered by active group member count.
func (handler Handler) popularGroupLists(ctx context.Context) ([]outsearch.ResultList, int, error) {
	if handler.GroupRooms == nil {
		return handler.roomList(ctx, "groups", nil)
	}
	groups, err := handler.GroupRooms.PopularGroups(ctx, handler.limit())
	if err != nil {
		return nil, 0, err
	}
	ids := make([]int64, 0, len(groups))
	for _, group := range groups {
		if group.Active() && group.HomeRoomID > 0 {
			ids = append(ids, group.HomeRoomID)
		}
	}
	rooms, err := handler.roomsByIDs(ctx, ids)
	if err != nil {
		return nil, 0, err
	}
	return handler.roomList(ctx, "groups", rooms)
}

// groupBaseLists builds one visible active group headquarters.
func (handler Handler) groupBaseLists(ctx context.Context, value string) ([]outsearch.ResultList, int, error) {
	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil || id <= 0 || handler.GroupRooms == nil {
		return handler.roomList(ctx, "group_base", nil)
	}
	group, found, err := handler.GroupRooms.Group(ctx, id)
	if err != nil {
		return nil, 0, err
	}
	if !found || !group.Active() {
		return handler.roomList(ctx, "group_base", nil)
	}
	rooms, err := handler.roomsByIDs(ctx, []int64{group.HomeRoomID})
	if err != nil {
		return nil, 0, err
	}
	return handler.roomList(ctx, "group_base", rooms)
}

// highestScoreLists builds durable score ranking.
func (handler Handler) highestScoreLists(ctx context.Context) ([]outsearch.ResultList, int, error) {
	rooms, err := handler.Rooms.ListHighestScore(ctx, handler.limit())
	if err != nil {
		return nil, 0, err
	}
	return handler.roomList(ctx, "highest_score", rooms)
}

// recommendedLists deterministically excludes owned and immediately recent rooms.
func (handler Handler) recommendedLists(ctx context.Context, playerID int64) ([]outsearch.ResultList, int, error) {
	popular, err := handler.Rooms.ListPopular(ctx, handler.limit())
	if err != nil {
		return nil, 0, err
	}
	recent, err := handler.Navigator.ListRecentRoomIDs(ctx, playerID, 10)
	if err != nil {
		return nil, 0, err
	}
	excluded := make(map[int64]struct{}, len(recent))
	for _, id := range recent {
		excluded[id] = struct{}{}
	}
	rooms := popular[:0]
	for _, room := range popular {
		if room.OwnerPlayerID != playerID {
			if _, found := excluded[room.ID]; !found {
				rooms = append(rooms, room)
			}
		}
	}
	return handler.roomList(ctx, "recommended", rooms)
}
