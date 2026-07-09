package http

import (
	"context"

	currencymodel "github.com/niflaot/pixels/internal/realm/inventory/currency/model"
	currencyservice "github.com/niflaot/pixels/internal/realm/inventory/currency/service"
	navmodel "github.com/niflaot/pixels/internal/realm/navigator/model"
	navservice "github.com/niflaot/pixels/internal/realm/navigator/service"
	playermodel "github.com/niflaot/pixels/internal/realm/player/model"
	playerservice "github.com/niflaot/pixels/internal/realm/player/service"
	roomlive "github.com/niflaot/pixels/internal/realm/room/live"
	roommodel "github.com/niflaot/pixels/internal/realm/room/model"
	roomservice "github.com/niflaot/pixels/internal/realm/room/service"
	netconn "github.com/niflaot/pixels/networking/connection"
	currencyroutes "github.com/niflaot/pixels/pkg/http/currency/routes"
	sharedmodel "github.com/niflaot/pixels/pkg/model"
	"go.uber.org/zap"
)

// testCurrencyDependencies composes HTTP currency route test dependencies.
func testCurrencyDependencies(registry *netconn.Registry, log *zap.Logger) currencyroutes.Dependencies {
	return currencyroutes.Dependencies{
		Finder: testFinder{}, Players: testPlayers(), Connections: registry,
		Currencies: testCurrencies(), Translations: testTranslations(), Log: log,
	}
}

// testCurrencies returns public client currency definitions.
func testCurrencies() testCurrencyReader {
	return testCurrencyReader{}
}

// testCurrencyReader returns HTTP test currency data.
type testCurrencyReader struct{}

// Wallet returns no HTTP test balances.
func (testCurrencyReader) Wallet(context.Context, int64) ([]currencymodel.Balance, error) {
	return nil, nil
}

// Balance returns a zero HTTP test balance.
func (testCurrencyReader) Balance(context.Context, int64, int32) (int64, error) {
	return 0, nil
}

// Types returns HTTP test currency definitions.
func (testCurrencyReader) Types(context.Context) ([]currencymodel.Definition, error) {
	return []currencymodel.Definition{{Type: -1, Key: "credits"}, {Type: 5, Key: "diamonds"}}, nil
}

// Grant returns a fake committed balance.
func (testCurrencyReader) Grant(_ context.Context, params currencyservice.GrantParams) (int64, error) {
	return params.Amount, nil
}

// Set returns a fake absolute balance.
func (testCurrencyReader) Set(_ context.Context, params currencyservice.SetParams) (int64, error) {
	return params.Amount, nil
}

// testFinder returns persistent test player records.
type testFinder struct{}

// FindByID finds a test player by id.
func (finder testFinder) FindByID(ctx context.Context, id int64) (playerservice.Record, bool, error) {
	if id != 2 {
		return playerservice.Record{}, false, nil
	}

	return testRecord(id), true, nil
}

// FindByUsername finds a test player by username.
func (finder testFinder) FindByUsername(context.Context, string) (playerservice.Record, bool, error) {
	return testRecord(2), true, nil
}

// testRecord returns a persistent test player record.
func testRecord(id int64) playerservice.Record {
	return playerservice.Record{
		Player: playermodel.Player{
			Base:     sharedmodel.Base{Identity: sharedmodel.Identity{ID: id}},
			Username: "test_player",
		},
		Profile: playermodel.Profile{
			PlayerID:        id,
			Look:            "hd-180-1",
			Gender:          playermodel.GenderMale,
			Motto:           "Test fixture.",
			AllowNameChange: true,
		},
	}
}

// testRooms returns an HTTP test room manager.
func testRooms() roomservice.Manager {
	return testRoomManager{}
}

// testRoomRuntime returns an HTTP test room runtime.
func testRoomRuntime() *roomlive.Registry {
	return roomlive.NewRegistry(nil)
}

// testNavigator returns an HTTP test navigator manager.
func testNavigator() navservice.Manager {
	return testNavigatorManager{}
}

// testRoomManager provides room data for HTTP tests.
type testRoomManager struct{}

// Create creates a test room.
func (manager testRoomManager) Create(context.Context, roomservice.CreateParams) (roommodel.Room, error) {
	return roommodel.Room{}, nil
}

// FindByID finds a test room by id.
func (manager testRoomManager) FindByID(context.Context, int64) (roommodel.Room, bool, error) {
	return testRoom(), true, nil
}

// ListByOwner lists test rooms by owner.
func (manager testRoomManager) ListByOwner(context.Context, int64) ([]roommodel.Room, error) {
	return []roommodel.Room{testRoom()}, nil
}

// ListPopular lists test rooms.
func (manager testRoomManager) ListPopular(context.Context, int) ([]roommodel.Room, error) {
	return []roommodel.Room{testRoom()}, nil
}

// ListHighestScore lists highest score test rooms.
func (manager testRoomManager) ListHighestScore(context.Context, int) ([]roommodel.Room, error) {
	return []roommodel.Room{testRoom()}, nil
}

// Search searches test rooms.
func (manager testRoomManager) Search(context.Context, string, int) ([]roommodel.Room, error) {
	return []roommodel.Room{testRoom()}, nil
}

// ListTags lists test room tags.
func (manager testRoomManager) ListTags(context.Context, int64) ([]roommodel.Tag, error) {
	return nil, nil
}

// SoftDelete soft deletes a test room.
func (manager testRoomManager) SoftDelete(context.Context, int64) error {
	return nil
}

// ListCategories lists test room categories.
func (manager testRoomManager) ListCategories(context.Context) ([]roommodel.Category, error) {
	return []roommodel.Category{{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 1}}, Caption: "Social", Visible: true}}, nil
}

// testRoom returns a test room.
func testRoom() roommodel.Room {
	return roommodel.Room{
		Base:          sharedmodel.Base{Identity: sharedmodel.Identity{ID: 1}},
		OwnerPlayerID: 2,
		OwnerName:     "test_player",
		Name:          "Test Room",
		ModelName:     "model_test",
		MaxUsers:      25,
	}
}

// testNavigatorManager provides navigator data for HTTP tests.
type testNavigatorManager struct{}

// AddFavorite adds a favorite room for tests.
func (manager testNavigatorManager) AddFavorite(context.Context, int64, int64) error { return nil }

// RemoveFavorite removes a favorite room for tests.
func (manager testNavigatorManager) RemoveFavorite(context.Context, int64, int64) error { return nil }

// ListFavoriteRoomIDs lists favorite room ids for tests.
func (manager testNavigatorManager) ListFavoriteRoomIDs(context.Context, int64) ([]int64, error) {
	return nil, nil
}

// SaveSearch saves a search for tests.
func (manager testNavigatorManager) SaveSearch(context.Context, navservice.SaveSearchParams) (navmodel.SavedSearch, error) {
	return navmodel.SavedSearch{}, nil
}

// DeleteSearch deletes a search for tests.
func (manager testNavigatorManager) DeleteSearch(context.Context, int64, int64) error { return nil }

// ListSavedSearches lists saved searches for tests.
func (manager testNavigatorManager) ListSavedSearches(context.Context, int64) ([]navmodel.SavedSearch, error) {
	return nil, nil
}

// SavePreference saves preferences for tests.
func (manager testNavigatorManager) SavePreference(context.Context, navmodel.Preference) (navmodel.Preference, error) {
	return navmodel.Preference{}, nil
}

// Preference returns navigator preferences for tests.
func (manager testNavigatorManager) Preference(context.Context, int64) (navmodel.Preference, error) {
	return navmodel.Preference{}, nil
}

// SaveCategoryPreference saves category preferences for tests.
func (manager testNavigatorManager) SaveCategoryPreference(context.Context, navmodel.CategoryPreference) (navmodel.CategoryPreference, error) {
	return navmodel.CategoryPreference{}, nil
}

// ListCategoryPreferences lists category preferences for tests.
func (manager testNavigatorManager) ListCategoryPreferences(context.Context, int64) ([]navmodel.CategoryPreference, error) {
	return nil, nil
}

// ListLiftedRooms lists lifted rooms for tests.
func (manager testNavigatorManager) ListLiftedRooms(context.Context) ([]navmodel.LiftedRoom, error) {
	return []navmodel.LiftedRoom{{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 1}}, RoomID: 1}}, nil
}
