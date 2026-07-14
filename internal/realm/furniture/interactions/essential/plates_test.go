package essential

import (
	"context"
	"testing"
	"time"

	playereffect "github.com/niflaot/pixels/internal/realm/player/effect"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	playermodel "github.com/niflaot/pixels/internal/realm/player/model"
	worldfurniture "github.com/niflaot/pixels/internal/realm/room/world/furniture"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
	netconn "github.com/niflaot/pixels/networking/connection"
)

// tileEffectManager records selected gender effects.
type tileEffectManager struct {
	// Manager supplies unused effect operations.
	playereffect.Manager
	// granted stores effect ids in call order.
	granted []int32
	// enabled stores selected effect ids in call order.
	enabled []int32
}

// Grant records one tile effect grant.
func (manager *tileEffectManager) Grant(_ context.Context, playerID int64, effectID int32, duration int32, source playereffect.Source) (playereffect.Effect, error) {
	manager.granted = append(manager.granted, effectID)
	return playereffect.Effect{PlayerID: playerID, ID: effectID, DurationSeconds: duration, RemainingCharges: 1}, nil
}

// Enable records one tile effect selection.
func (manager *tileEffectManager) Enable(_ context.Context, _ int64, effectID int32) error {
	manager.enabled = append(manager.enabled, effectID)
	return nil
}

// TestEffectTileUsesPlayerGender verifies movement grants the configured variant.
func TestEffectTileUsesPlayerGender(t *testing.T) {
	maleEffect, femaleEffect := int32(201), int32(202)
	manager := &tileEffectManager{}
	players := playerlive.NewRegistry()
	for id, gender := range []playermodel.Gender{playermodel.GenderMale, playermodel.GenderFemale} {
		peer, _ := playerlive.NewSessionPeer(netconn.ID(string(rune('a'+id))), "websocket", time.Now())
		player, _ := playerlive.NewPlayer(playerlive.Snapshot{ID: int64(id + 1), Username: string(rune('a' + id)), Gender: gender}, peer)
		_ = players.Add(player)
	}
	service := &Service{effects: manager, players: players}
	item := worldfurniture.Item{Definition: worldfurniture.Definition{EffectMale: &maleEffect, EffectFemale: &femaleEffect}}
	if err := service.giveTileEffect(context.Background(), 1, item); err != nil {
		t.Fatal(err)
	}
	if err := service.giveTileEffect(context.Background(), 2, item); err != nil {
		t.Fatal(err)
	}
	if len(manager.granted) != 2 || manager.granted[0] != maleEffect || manager.granted[1] != femaleEffect || manager.enabled[0] != maleEffect || manager.enabled[1] != femaleEffect {
		t.Fatalf("granted=%v enabled=%v", manager.granted, manager.enabled)
	}
}

// TestPressurePlateDerivesOccupancy verifies debounced walk-on and walk-off state.
func TestPressurePlateDerivesOccupancy(t *testing.T) {
	item := essentialItem("pressureplate", 2)
	item.Definition.AllowWalk, item.Definition.StackHeight = true, 0
	active := essentialRoom(t, item, 1)
	service := &Service{}
	if _, err := active.TeleportUnit(1, item.Point, worldunit.RotationEast, false); err != nil {
		t.Fatalf("position unit: %v", err)
	}
	service.schedulePressure(context.Background(), active, item)
	active.RunScheduled(time.Now().Add(time.Second))
	pressed, _ := active.FurnitureItem(item.ID)
	if pressed.ExtraData != "1" {
		t.Fatalf("expected pressed state, got %q", pressed.ExtraData)
	}
	if _, err := active.TeleportUnit(1, grid.MustPoint(0, 0), worldunit.RotationEast, false); err != nil {
		t.Fatalf("move unit off: %v", err)
	}
	service.schedulePressure(context.Background(), active, pressed)
	active.RunScheduled(time.Now().Add(time.Second))
	released, _ := active.FurnitureItem(item.ID)
	if released.ExtraData != "0" {
		t.Fatalf("expected released state, got %q", released.ExtraData)
	}
}

// TestColorPlateClampsDeltas verifies bounded occupancy state changes.
func TestColorPlateClampsDeltas(t *testing.T) {
	item := essentialItem("colorplate", 3)
	active := essentialRoom(t, item, 1)
	service := &Service{}
	for range 4 {
		current, _ := active.FurnitureItem(item.ID)
		if err := service.changeColorPlate(context.Background(), active, current, 1); err != nil {
			t.Fatalf("increment plate: %v", err)
		}
	}
	current, _ := active.FurnitureItem(item.ID)
	if current.ExtraData != "2" {
		t.Fatalf("expected clamped high state, got %q", current.ExtraData)
	}
	for range 4 {
		current, _ = active.FurnitureItem(item.ID)
		_ = service.changeColorPlate(context.Background(), active, current, -1)
	}
	current, _ = active.FurnitureItem(item.ID)
	if current.ExtraData != "0" {
		t.Fatalf("expected clamped low state, got %q", current.ExtraData)
	}
}

// BenchmarkColorPlate measures a movement-driven plate state update.
func BenchmarkColorPlate(b *testing.B) {
	item := essentialItem("colorplate", 5)
	active := essentialRoom(b, item, 1)
	service := &Service{}
	ctx := context.Background()
	b.ReportAllocs()
	for b.Loop() {
		current, _ := active.FurnitureItem(item.ID)
		_ = service.changeColorPlate(ctx, active, current, 1)
		_, _ = active.SetFurnitureExtraData(item.ID, "0")
	}
}
