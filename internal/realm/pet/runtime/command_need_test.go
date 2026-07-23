package runtime

import (
	"context"
	"errors"
	"testing"
	"time"

	petrecord "github.com/niflaot/pixels/internal/realm/pet/record"
	petwork "github.com/niflaot/pixels/internal/realm/pet/runtime/work"
	worldfurniture "github.com/niflaot/pixels/internal/realm/room/world/furniture"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
)

// commandNeedConsumer records one contextual consumption request.
type commandNeedConsumer struct {
	// calls receives requested item identifiers.
	calls chan int64
}

// ConsumeNeed records one contextual consumption request.
func (consumer *commandNeedConsumer) ConsumeNeed(_ context.Context, _ int64, _ int64, itemID int64) error {
	consumer.calls <- itemID
	return nil
}

// TestCommandNeedRequiresMatchingFurniture verifies commands do not grant stats without a product.
func TestCommandNeedRequiresMatchingFurniture(t *testing.T) {
	service, active, _ := unitLookupFixture(t)
	store := &actionStore{updated: true}
	service.store = store
	service.needs = &commandNeedConsumer{calls: make(chan int64, 1)}
	err := service.ExecuteAction(context.Background(), active.ID(), 50, 1, CommandAction{ID: 43, Mode: ActionNeed, Need: CommandNeedFood}, petrecord.Command{RequiredLevel: 1, ExperienceReward: 5})
	controller, _ := service.Active(active.ID(), 50)
	controller.mutex.Lock()
	pending := controller.commandNeed.itemID
	controller.mutex.Unlock()
	_, _, _ = service.rooms.Close(context.Background(), active.ID())
	if !errors.Is(err, petrecord.ErrInvalidProduct) || store.calls != 0 || pending != 0 {
		t.Fatalf("err=%v stat_calls=%d pending=%d", err, store.calls, pending)
	}
}

// TestCommandNeedWalksThenConsumes verifies contextual actions wait for physical arrival.
func TestCommandNeedWalksThenConsumes(t *testing.T) {
	service, active, _ := unitLookupFixture(t)
	t.Cleanup(func() { _, _, _ = service.rooms.Close(context.Background(), active.ID()) })
	store := &actionStore{updated: true}
	consumer := &commandNeedConsumer{calls: make(chan int64, 1)}
	service.store = store
	service.needs = consumer
	service.work = petwork.New(8, 1, nil)
	service.Start()
	t.Cleanup(service.Stop)
	item := worldfurniture.Item{ID: 99, Definition: worldfurniture.Definition{InteractionType: "pet_food", Width: 1, Length: 1, AllowWalk: true}, Point: grid.MustPoint(2, 0)}
	if _, err := active.ReloadFurniture(item.ID, &item); err != nil {
		t.Fatal(err)
	}
	if err := service.ExecuteAction(context.Background(), active.ID(), 50, 1, CommandAction{ID: 43, Mode: ActionNeed, Need: CommandNeedFood}, petrecord.Command{RequiredLevel: 1, ExperienceReward: 5}); err != nil {
		t.Fatal(err)
	}
	if store.calls != 0 {
		t.Fatalf("command mutated stats before consumption: %d", store.calls)
	}
	unit, _ := active.UnitMotion(EntityKey(50))
	if !unit.Moving {
		t.Fatalf("pet did not start contextual movement: %+v", unit)
	}
	for range 3 {
		active.Tick()
		if err := service.Cycle(context.Background(), active, time.Now()); err != nil {
			t.Fatal(err)
		}
	}
	select {
	case itemID := <-consumer.calls:
		if itemID != item.ID {
			t.Fatalf("consumed item=%d", itemID)
		}
	case <-time.After(time.Second):
		t.Fatal("contextual consumption was not dispatched")
	}
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		active.RunScheduled(time.Now().Add(time.Second))
		controller, _ := service.Active(active.ID(), 50)
		controller.mutex.Lock()
		pending := controller.commandNeed.itemID != 0 || controller.needPending
		controller.mutex.Unlock()
		if !pending {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("contextual command did not finish")
}

// TestCommandNeedCancelsWhenFurnitureDisappears verifies pickup during walking cannot consume stale state.
func TestCommandNeedCancelsWhenFurnitureDisappears(t *testing.T) {
	service, active, _ := unitLookupFixture(t)
	service.needs = &commandNeedConsumer{calls: make(chan int64, 1)}
	item := worldfurniture.Item{ID: 99, Definition: worldfurniture.Definition{InteractionType: "pet_drink", Width: 1, Length: 1, AllowWalk: true}, Point: grid.MustPoint(2, 0)}
	if _, err := active.ReloadFurniture(item.ID, &item); err != nil {
		t.Fatal(err)
	}
	if err := service.ExecuteAction(context.Background(), active.ID(), 50, 1, CommandAction{ID: 14, Mode: ActionNeed, Need: CommandNeedDrink}, petrecord.Command{RequiredLevel: 1}); err != nil {
		t.Fatal(err)
	}
	if _, err := active.ReloadFurniture(item.ID, nil); err != nil {
		t.Fatal(err)
	}
	controller, _ := service.Active(active.ID(), 50)
	service.resolveCommandNeed(context.Background(), active, controller, petrecord.Pet{ID: 50})
	controller.mutex.Lock()
	pending := controller.commandNeed.itemID
	controller.mutex.Unlock()
	_, _, _ = service.rooms.Close(context.Background(), active.ID())
	if pending != 0 {
		t.Fatalf("stale contextual product remained pending: %d", pending)
	}
}

// TestCommandNeedReleasesSaturatedDispatch verifies a rejected worker job cannot strand the pet.
func TestCommandNeedReleasesSaturatedDispatch(t *testing.T) {
	service, active, _ := unitLookupFixture(t)
	service.needs = &commandNeedConsumer{calls: make(chan int64, 1)}
	controller, _ := service.Active(active.ID(), 50)
	pending := commandNeedState{itemID: 99, actionID: 43, kind: CommandNeedFood}
	controller.mutex.Lock()
	controller.commandNeed = pending
	controller.needPending = true
	controller.mutex.Unlock()
	service.dispatchCommandNeed(active, controller, petrecord.Pet{ID: 50}, pending)
	controller.mutex.Lock()
	stillPending := controller.commandNeed.itemID != 0 || controller.needPending
	controller.mutex.Unlock()
	_, _, _ = service.rooms.Close(context.Background(), active.ID())
	if stillPending {
		t.Fatal("rejected dispatch stranded contextual state")
	}
}

// TestCommandNeedStopsWhenFeedingIsDisabled verifies room policy changes cancel an active walk.
func TestCommandNeedStopsWhenFeedingIsDisabled(t *testing.T) {
	service, active, _ := unitLookupFixture(t)
	service.needs = &commandNeedConsumer{calls: make(chan int64, 1)}
	item := worldfurniture.Item{ID: 99, Definition: worldfurniture.Definition{InteractionType: "pet_food", Width: 1, Length: 1, AllowWalk: true}, Point: grid.MustPoint(2, 0)}
	if _, err := active.ReloadFurniture(item.ID, &item); err != nil {
		t.Fatal(err)
	}
	if err := service.ExecuteAction(context.Background(), active.ID(), 50, 1, CommandAction{ID: 43, Mode: ActionNeed, Need: CommandNeedFood}, petrecord.Command{RequiredLevel: 1}); err != nil {
		t.Fatal(err)
	}
	active.UpdatePetSettings(true, false)
	controller, _ := service.Active(active.ID(), 50)
	service.resolveCommandNeed(context.Background(), active, controller, petrecord.Pet{ID: 50})
	controller.mutex.Lock()
	pending := controller.commandNeed.itemID
	controller.mutex.Unlock()
	_, _, _ = service.rooms.Close(context.Background(), active.ID())
	if pending != 0 {
		t.Fatalf("disabled feeding left contextual state pending: %d", pending)
	}
}
