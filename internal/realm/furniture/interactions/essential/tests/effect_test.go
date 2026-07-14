package tests

import (
	"context"
	"testing"

	"github.com/niflaot/pixels/internal/realm/furniture/interactions/essential"
	playereffect "github.com/niflaot/pixels/internal/realm/player/effect"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	worldfurniture "github.com/niflaot/pixels/internal/realm/room/world/furniture"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldpath "github.com/niflaot/pixels/internal/realm/room/world/path"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
	netconn "github.com/niflaot/pixels/networking/connection"
)

// effectManager records furniture-driven effect behavior.
type effectManager struct {
	// granted stores the granted effect id.
	granted int32
	// enabled stores the enabled effect id.
	enabled int32
}

// List returns no fixture effects.
func (*effectManager) List(context.Context, int64) ([]playereffect.Effect, error) { return nil, nil }

// Grant records one furniture effect grant.
func (manager *effectManager) Grant(_ context.Context, playerID int64, effectID int32, duration int32, source playereffect.Source) (playereffect.Effect, error) {
	manager.granted = effectID
	return playereffect.Effect{PlayerID: playerID, ID: effectID, DurationSeconds: duration, RemainingCharges: 1}, nil
}

// Enable records one immediate selection.
func (manager *effectManager) Enable(_ context.Context, _ int64, effectID int32) error {
	manager.enabled = effectID
	return nil
}

// Activate returns an unused fixture effect.
func (*effectManager) Activate(context.Context, int64, int32) (playereffect.Effect, error) {
	return playereffect.Effect{}, nil
}

// Revoke accepts an unused fixture revocation.
func (*effectManager) Revoke(context.Context, int64, int32) error { return nil }

// TestEffectGiverGrantsAndEnables verifies the furniture effect pipeline.
func TestEffectGiverGrantsAndEnables(t *testing.T) {
	manager := &effectManager{}
	service := essential.NewWithEffects(nil, nil, nil, nil, nil, nil, manager, nil, nil)
	point, _ := grid.NewPoint(1, 1)
	itemPoint, _ := grid.NewPoint(2, 1)
	item := worldfurniture.Item{ID: 1, Point: itemPoint, Definition: worldfurniture.Definition{Width: 1, Length: 1, InteractionType: "effect_giver", EffectPool: []int32{101}}}
	room, err := roomlive.NewRoom(roomlive.Snapshot{ID: 9, OwnerPlayerID: 7, MaxUsers: 25})
	if err != nil {
		t.Fatal(err)
	}
	roomGrid, err := grid.Parse("000\r000\r000")
	if err != nil {
		t.Fatal(err)
	}
	if err = room.LoadWorld(roomlive.WorldConfig{Grid: roomGrid, Door: worldpath.Position{Point: point}, Body: worldunit.RotationSouth, Head: worldunit.RotationSouth}); err != nil {
		t.Fatal(err)
	}
	if _, err = room.Join(roomlive.Occupant{PlayerID: 7, ConnectionID: netconn.ID("effect"), ConnectionKind: netconn.Kind("websocket")}); err != nil {
		t.Fatal(err)
	}
	handled, err := service.Use(context.Background(), essential.Request{PlayerID: 7, Room: room, Item: item})
	if err != nil || !handled || manager.granted != 101 || manager.enabled != 101 {
		t.Fatalf("handled=%t granted=%d enabled=%d err=%v", handled, manager.granted, manager.enabled, err)
	}
}
