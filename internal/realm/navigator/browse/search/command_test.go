package search

import (
	"context"
	"testing"

	grouprecord "github.com/niflaot/pixels/internal/realm/group/record"
	roommodel "github.com/niflaot/pixels/internal/realm/room/record/model"
	roomservice "github.com/niflaot/pixels/internal/realm/room/record/service"
	"github.com/niflaot/pixels/networking/outbound/navigator/browse/roomcard"
	outsearch "github.com/niflaot/pixels/networking/outbound/navigator/browse/searchresult"
	sharedmodel "github.com/niflaot/pixels/pkg/model"
)

// TestCardsMapsRoomRecords verifies search card projection.
func TestCardsMapsRoomRecords(t *testing.T) {
	cards, err := Handler{}.cards(context.Background(), []roommodel.Room{{
		Base:          sharedmodel.Base{Identity: sharedmodel.Identity{ID: 9}},
		OwnerPlayerID: 7,
		OwnerName:     "demo",
		Name:          "Demo Room",
		MaxUsers:      25,
	}})

	if err != nil || len(cards) != 1 || cards[0].RoomID != 9 || cards[0].Ranking != 1 {
		t.Fatalf("unexpected cards %#v", cards)
	}
}

// TestResultListsMapsRoomAdsView verifies the events tab uses category lists.
func TestResultListsMapsRoomAdsView(t *testing.T) {
	lists, count, err := Handler{Rooms: testRooms{}}.resultLists(context.Background(), 1, "roomads_view", "")
	if err != nil {
		t.Fatalf("build lists: %v", err)
	}
	if count != 1 || len(lists) != 1 || lists[0].Code != "categories" {
		t.Fatalf("unexpected lists %#v count %d", lists, count)
	}
}

// TestResultListsMapsPopularGroups verifies the group profile link returns only ranked headquarters.
func TestResultListsMapsPopularGroups(t *testing.T) {
	rooms := testRooms{rooms: map[int64]roommodel.Room{
		31: {Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 31}}, Name: "Large Group HQ", MaxUsers: 25},
		32: {Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 32}}, Name: "Small Group HQ", MaxUsers: 25},
	}}
	groups := testGroupRooms{popular: []grouprecord.Group{
		{ID: 9, HomeRoomID: 31, MemberCount: 40},
		{ID: 8, HomeRoomID: 32, MemberCount: 3},
	}}
	lists, count, err := (Handler{Rooms: rooms, GroupRooms: groups}).resultLists(context.Background(), 1, "groups", "")
	if err != nil {
		t.Fatalf("build lists: %v", err)
	}
	if count != 2 || len(lists) != 1 || lists[0].Code != "groups" || lists[0].Rooms[0].RoomID != 31 || lists[0].Rooms[1].RoomID != 32 {
		t.Fatalf("unexpected lists %#v count %d", lists, count)
	}
}

// TestVisibleRoomIDsExtractsResultRooms verifies visible room id snapshots.
func TestVisibleRoomIDsExtractsResultRooms(t *testing.T) {
	roomIDs := VisibleRoomIDs([]outsearch.ResultList{
		{Rooms: []roomcard.Card{{RoomID: 4}, {RoomID: 0}}},
		{Rooms: []roomcard.Card{{RoomID: 7}}},
	})

	if len(roomIDs) != 2 || roomIDs[0] != 4 || roomIDs[1] != 7 {
		t.Fatalf("unexpected room ids %#v", roomIDs)
	}
}

// TestFavoriteVisibleProtectsInvisibleRooms verifies owner and rights visibility.
func TestFavoriteVisibleProtectsInvisibleRooms(t *testing.T) {
	room := roommodel.Room{
		Base:          sharedmodel.Base{Identity: sharedmodel.Identity{ID: 9}},
		OwnerPlayerID: 7, DoorMode: roommodel.DoorModeInvisible,
	}
	if favoriteVisible(8, room, nil) {
		t.Fatal("expected hidden guest favorite")
	}
	if !favoriteVisible(7, room, nil) {
		t.Fatal("expected owner favorite")
	}
	if !favoriteVisible(8, room, []int64{9}) {
		t.Fatal("expected rights favorite")
	}
}

// testRooms provides room data for search tests.
type testRooms struct {
	// rooms stores optional room fixtures by identifier.
	rooms map[int64]roommodel.Room
}

// Create creates a room for search tests.
func (testRooms) Create(context.Context, roomservice.CreateParams) (roommodel.Room, error) {
	return roommodel.Room{}, nil
}

// FindByID finds a room for search tests.
func (rooms testRooms) FindByID(_ context.Context, id int64) (roommodel.Room, bool, error) {
	room, found := rooms.rooms[id]
	return room, found, nil
}

// ListByOwner lists owned rooms for search tests.
func (testRooms) ListByOwner(context.Context, int64) ([]roommodel.Room, error) {
	return nil, nil
}

// ListPopular lists popular rooms for search tests.
func (testRooms) ListPopular(context.Context, int) ([]roommodel.Room, error) {
	return nil, nil
}

// ListHighestScore lists highest score rooms for search tests.
func (testRooms) ListHighestScore(context.Context, int) ([]roommodel.Room, error) {
	return []roommodel.Room{{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 3}}, Name: "Event Room", OwnerName: "demo", MaxUsers: 10}}, nil
}

// Search searches rooms for search tests.
func (testRooms) Search(context.Context, string, int) ([]roommodel.Room, error) {
	return nil, nil
}

// ListTags lists room tags for search tests.
func (testRooms) ListTags(context.Context, int64) ([]roommodel.Tag, error) {
	return nil, nil
}

// SoftDelete deletes rooms for search tests.
func (testRooms) SoftDelete(context.Context, int64) error {
	return nil
}

// ListCategories lists categories for search tests.
func (testRooms) ListCategories(context.Context) ([]roommodel.Category, error) {
	return nil, nil
}

// testGroupRooms provides group headquarters for search tests.
type testGroupRooms struct {
	// popular stores ranked group fixtures.
	popular []grouprecord.Group
}

// PlayerGroups returns no membership fixtures.
func (testGroupRooms) PlayerGroups(context.Context, int64) ([]grouprecord.PlayerGroup, error) {
	return nil, nil
}

// PopularGroups returns ranked group fixtures.
func (groups testGroupRooms) PopularGroups(context.Context, int) ([]grouprecord.Group, error) {
	return groups.popular, nil
}

// Group returns no single group fixture.
func (testGroupRooms) Group(context.Context, int64) (grouprecord.Group, bool, error) {
	return grouprecord.Group{}, false, nil
}
