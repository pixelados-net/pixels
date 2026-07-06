package search

import (
	"context"
	"testing"

	roommodel "github.com/niflaot/pixels/internal/realm/room/model"
	roomservice "github.com/niflaot/pixels/internal/realm/room/service"
	sharedmodel "github.com/niflaot/pixels/pkg/model"
)

// TestCardsMapsRoomRecords verifies search card projection.
func TestCardsMapsRoomRecords(t *testing.T) {
	cards := Handler{}.cards([]roommodel.Room{{
		Base:          sharedmodel.Base{Identity: sharedmodel.Identity{ID: 9}},
		OwnerPlayerID: 7,
		OwnerName:     "demo",
		Name:          "Demo Room",
		MaxUsers:      25,
	}})

	if len(cards) != 1 || cards[0].RoomID != 9 || cards[0].Ranking != 1 {
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

// testRooms provides room data for search tests.
type testRooms struct{}

// Create creates a room for search tests.
func (testRooms) Create(context.Context, roomservice.CreateParams) (roommodel.Room, error) {
	return roommodel.Room{}, nil
}

// FindByID finds a room for search tests.
func (testRooms) FindByID(context.Context, int64) (roommodel.Room, bool, error) {
	return roommodel.Room{}, false, nil
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
