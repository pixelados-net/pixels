package lifecycle

import (
	"context"
	"errors"
	"testing"

	botbehavior "github.com/niflaot/pixels/internal/realm/bot/behavior"
	botcore "github.com/niflaot/pixels/internal/realm/bot/core"
	botpolicy "github.com/niflaot/pixels/internal/realm/bot/policy"
	botrecord "github.com/niflaot/pixels/internal/realm/bot/record"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldpath "github.com/niflaot/pixels/internal/realm/room/world/path"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
	"go.uber.org/zap"
)

// lifecycleStore provides mutable records required by lifecycle tests.
type lifecycleStore struct {
	botrecord.Store
	// records stores bots by durable id.
	records map[int64]botrecord.Bot
	// inventoryCount overrides the inventory size query.
	inventoryCount int
}

// CountInventory returns the configured inventory size.
func (store *lifecycleStore) CountInventory(context.Context, int64) (int, error) {
	return store.inventoryCount, nil
}

// Room returns bots placed in the requested room.
func (store *lifecycleStore) Room(_ context.Context, roomID int64) ([]botrecord.Bot, error) {
	result := make([]botrecord.Bot, 0)
	for _, bot := range store.records {
		if bot.RoomID != nil && *bot.RoomID == roomID {
			result = append(result, bot)
		}
	}
	return result, nil
}

// Find returns one configured record.
func (store *lifecycleStore) Find(_ context.Context, botID int64) (botrecord.Bot, bool, error) {
	bot, found := store.records[botID]
	return bot, found, nil
}

// Inventory returns every configured inventory bot for one owner.
func (store *lifecycleStore) Inventory(_ context.Context, playerID int64) ([]botrecord.Bot, error) {
	result := make([]botrecord.Bot, 0)
	for _, bot := range store.records {
		if bot.OwnerPlayerID == playerID && bot.Inventory() {
			result = append(result, bot)
		}
	}
	return result, nil
}

// Place persists one test placement.
func (store *lifecycleStore) Place(_ context.Context, botID int64, ownerID int64, roomID int64, x int, y int, z float64, rotation int16) (botrecord.Bot, bool, error) {
	bot, found := store.records[botID]
	if !found || bot.OwnerPlayerID != ownerID || !bot.Inventory() {
		return botrecord.Bot{}, false, nil
	}
	bot.RoomID, bot.X, bot.Y, bot.Z, bot.Rotation = &roomID, &x, &y, &z, &rotation
	store.records[botID] = bot
	return bot, true, nil
}

// Pickup returns one placed bot to the requested owner.
func (store *lifecycleStore) Pickup(_ context.Context, botID int64, roomID int64, ownerID int64) (botrecord.Bot, bool, error) {
	bot, found := store.records[botID]
	if !found || bot.RoomID == nil || *bot.RoomID != roomID {
		return botrecord.Bot{}, false, nil
	}
	bot.OwnerPlayerID, bot.RoomID, bot.X, bot.Y, bot.Z, bot.Rotation = ownerID, nil, nil, nil, nil, nil
	store.records[botID] = bot
	return bot, true, nil
}

// ForcePickup returns one placed bot to its existing owner.
func (store *lifecycleStore) ForcePickup(ctx context.Context, botID int64) (botrecord.Bot, bool, error) {
	bot, found := store.records[botID]
	if !found || bot.RoomID == nil {
		return botrecord.Bot{}, false, nil
	}
	return store.Pickup(ctx, botID, *bot.RoomID, bot.OwnerPlayerID)
}

// Delete removes one owned inventory bot.
func (store *lifecycleStore) Delete(_ context.Context, botID int64, ownerID int64) (bool, error) {
	bot, found := store.records[botID]
	if !found || bot.OwnerPlayerID != ownerID || !bot.Inventory() {
		return false, nil
	}
	delete(store.records, botID)
	return true, nil
}

// TestPlaceRejectsPlayerAndBotOccupiedTiles verifies both collision paths.
func TestPlaceRejectsPlayerAndBotOccupiedTiles(t *testing.T) {
	store := &lifecycleStore{records: map[int64]botrecord.Bot{1: inventoryBot(1), 2: placedBot(2, 1, 0)}}
	service, room := lifecycleFixture(t, botpolicy.Config{MaxPerRoom: 10}, store)
	if _, err := room.Join(roomlive.Occupant{PlayerID: 1, ConnectionID: "1", ConnectionKind: "test"}); err != nil {
		t.Fatalf("join: %v", err)
	}
	_, err := service.Place(context.Background(), PlaceParams{BotID: 1, ActorPlayerID: 1, RoomID: 9, Point: grid.MustPoint(0, 0)})
	if !errors.Is(err, botrecord.ErrTileNotFree) {
		t.Fatalf("player tile error=%v", err)
	}
	_, err = service.Place(context.Background(), PlaceParams{BotID: 1, ActorPlayerID: 1, RoomID: 9, Point: grid.MustPoint(1, 0)})
	if !errors.Is(err, botrecord.ErrTileNotFree) {
		t.Fatalf("bot tile error=%v", err)
	}
}

// TestPlaceEnforcesLoadedRoomLimit verifies persisted bots count before placement.
func TestPlaceEnforcesLoadedRoomLimit(t *testing.T) {
	store := &lifecycleStore{records: map[int64]botrecord.Bot{1: placedBot(1, 1, 0), 2: inventoryBot(2)}}
	service, _ := lifecycleFixture(t, botpolicy.Config{MaxPerRoom: 1}, store)
	_, err := service.Place(context.Background(), PlaceParams{BotID: 2, ActorPlayerID: 1, RoomID: 9, Point: grid.MustPoint(2, 0)})
	if !errors.Is(err, botrecord.ErrRoomLimit) {
		t.Fatalf("limit error=%v", err)
	}
}

// TestPickupEnforcesInventoryLimit verifies the native 25-bot bound.
func TestPickupEnforcesInventoryLimit(t *testing.T) {
	store := &lifecycleStore{records: map[int64]botrecord.Bot{1: placedBot(1, 1, 0)}, inventoryCount: 25}
	service, _ := lifecycleFixture(t, botpolicy.Config{MaxInventory: 25}, store)
	_, err := service.Pickup(context.Background(), PickupParams{BotID: 1, ActorPlayerID: 1, RoomID: 9})
	if !errors.Is(err, botrecord.ErrInventoryLimit) {
		t.Fatalf("inventory limit error=%v", err)
	}
}

// TestLifecycleHappyPaths verifies inventory, placement, pickup, force pickup, and deletion.
func TestLifecycleHappyPaths(t *testing.T) {
	store := &lifecycleStore{records: map[int64]botrecord.Bot{1: inventoryBot(1), 2: placedBot(2, 2, 0)}}
	service, room := lifecycleFixture(t, botpolicy.Config{MaxPerRoom: 10, MaxInventory: 25}, store)
	items, err := service.Inventory(context.Background(), 1)
	if err != nil || len(items) != 1 {
		t.Fatalf("inventory=%v err=%v", items, err)
	}
	if _, found, err := service.Find(context.Background(), 1); err != nil || !found {
		t.Fatalf("find found=%t err=%v", found, err)
	}
	placed, err := service.Place(context.Background(), PlaceParams{BotID: 1, ActorPlayerID: 1, RoomID: 9, Point: grid.MustPoint(1, 0)})
	if err != nil || placed.RoomID == nil {
		t.Fatalf("place=%#v err=%v", placed, err)
	}
	picked, err := service.Pickup(context.Background(), PickupParams{BotID: 1, ActorPlayerID: 1, RoomID: 9})
	if err != nil || !picked.Inventory() {
		t.Fatalf("pickup=%#v err=%v", picked, err)
	}
	forced, err := service.ForcePickup(context.Background(), 2)
	if err != nil || !forced.Inventory() {
		t.Fatalf("force=%#v err=%v", forced, err)
	}
	if err = service.Delete(context.Background(), 1, 1); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, found := room.Unit(botcore.EntityKey(1)); found {
		t.Fatal("expected picked bot removed from room")
	}
}

// lifecycleFixture creates bot lifecycle state over a loaded owner room.
func lifecycleFixture(t testing.TB, config botpolicy.Config, store *lifecycleStore) (*Service, *roomlive.Room) {
	t.Helper()
	rooms := roomlive.NewRegistry(nil)
	active, err := rooms.Activate(roomlive.Snapshot{ID: 9, OwnerPlayerID: 1, MaxUsers: 100})
	if err != nil {
		t.Fatalf("activate: %v", err)
	}
	t.Cleanup(func() { active.Close() })
	roomGrid, err := grid.Parse("000", grid.WithDoor(0, 0))
	if err != nil {
		t.Fatalf("grid: %v", err)
	}
	if err = active.LoadWorld(roomlive.WorldConfig{Grid: roomGrid, Door: worldpath.Position{Point: grid.MustPoint(0, 0)}, Body: worldunit.RotationSouth, Head: worldunit.RotationSouth}); err != nil {
		t.Fatalf("world: %v", err)
	}
	behaviors := botbehavior.NewRegistry()
	if err = botbehavior.RegisterBuiltins(behaviors); err != nil {
		t.Fatalf("behaviors: %v", err)
	}
	runtime := botcore.New(config, store, rooms, nil, playerlive.NewRegistry(), nil, nil, behaviors, nil, nil, nil, nil, nil, zap.NewNop())
	return New(config, store, rooms, nil, runtime), active
}

// inventoryBot creates one valid inventory record.
func inventoryBot(id int64) botrecord.Bot {
	return botrecord.Bot{ID: id, OwnerPlayerID: 1, BehaviorType: botrecord.BehaviorGeneric, Name: "Bot", Figure: "hd-180-1", Gender: "M", ChatDelaySeconds: 10}
}

// placedBot creates one valid placed record.
func placedBot(id int64, x int, y int) botrecord.Bot {
	bot := inventoryBot(id)
	roomID, z, rotation := int64(9), float64(0), int16(2)
	bot.RoomID, bot.X, bot.Y, bot.Z, bot.Rotation = &roomID, &x, &y, &z, &rotation
	return bot
}
